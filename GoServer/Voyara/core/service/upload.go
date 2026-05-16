package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	maxImageSize  = 1 * 1024 * 1024 // 1MB
	maxImageCount = 5
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func ValidateImageFile(filename string, data []byte) error {
	if len(data) > maxImageSize {
		return fmt.Errorf("image too large: %d bytes (max %d)", len(data), maxImageSize)
	}
	ext := strings.ToLower(path.Ext(filename))
	if !allowedImageExts[ext] {
		return fmt.Errorf("file type %s not allowed (jpg, jpeg, png, webp only)", ext)
	}
	return nil
}

// UploadToS3 uploads a file to S3 and returns the public URL.
// Uses EC2 IAM Role or environment variables for AWS credentials.
// AWS_S3_PREFIX can be set to a subfolder prefix (e.g. "dev" -> "dev/voyara/images/uuid.ext").
// TODO: For production, use CloudFront or presigned URLs instead of public-read.
func UploadToS3(data []byte, filename string) (string, error) {
	bucket := os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("AWS_S3_BUCKET not set")
	}
	region := os.Getenv("AWS_S3_REGION")
	if region == "" {
		region = "ap-southeast-2"
	}

	ext := strings.ToLower(path.Ext(filename))
	uuidStr, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("generate uuid: %v", err)
	}
	prefix := os.Getenv("AWS_S3_PREFIX")
	key := fmt.Sprintf("voyara/images/%s%s", uuidStr, ext)
	if prefix != "" {
		prefix = strings.TrimSuffix(prefix, "/")
		key = prefix + "/" + key
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	// If no explicit keys, try loading from EC2 metadata or shared config
	if accessKey == "" || secretKey == "" {
		creds, err := getCredentialsFromConfig()
		if err != nil {
			return "", fmt.Errorf("no AWS credentials available: %v", err)
		}
		accessKey = creds["accessKey"]
		secretKey = creds["secretKey"]
		sessionToken = creds["sessionToken"]
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	body := bytes.NewReader(data)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return "", fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Host", fmt.Sprintf("%s.s3.%s.amazonaws.com", bucket, region))
	req.Header.Set("Content-Type", detectContentType(ext))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	req.ContentLength = int64(len(data))

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	req.Header.Set("X-Amz-Date", amzDate)

	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// AWS Signature V4
	signS3V4(req, region, "s3", accessKey, secretKey, sessionToken, data)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("s3 upload request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("s3 upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	return publicURL, nil
}

func detectContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// getCredentialsFromConfig uses the existing AWS SDK config to load credentials.
func getCredentialsFromConfig() (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("retrieve aws credentials: %v", err)
	}
	result := map[string]string{
		"accessKey":    creds.AccessKeyID,
		"secretKey":    creds.SecretAccessKey,
		"sessionToken": creds.SessionToken,
	}
	return result, nil
}

// ── AWS SigV4 signing ──

func signS3V4(req *http.Request, region, service, accessKey, secretKey, sessionToken string, body []byte) {
	now, _ := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))
	dateStr := now.Format("20060102")

	bodyHash := sha256Hex(body)
	req.Header.Set("x-amz-content-sha256", bodyHash)

	canonicalURI := req.URL.Path
	canonicalQuery := req.URL.RawQuery

	orderedHeaders := []string{"content-length", "content-type", "host", "x-amz-content-sha256", "x-amz-date"}
	if sessionToken != "" {
		orderedHeaders = append(orderedHeaders, "x-amz-security-token")
	}

	var headers []string
	for _, k := range orderedHeaders {
		if v := req.Header.Get(k); v != "" {
			headers = append(headers, k+":"+strings.TrimSpace(v))
		}
	}

	canonicalHeaders := strings.Join(headers, "\n") + "\n"
	signedHeadersStr := strings.Join(orderedHeaders, ";")

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method, canonicalURI, canonicalQuery, canonicalHeaders, signedHeadersStr, bodyHash)

	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStr, region, service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, req.Header.Get("X-Amz-Date"), credentialScope, sha256Hex([]byte(canonicalRequest)))

	signingKey := getSignatureKey(secretKey, dateStr, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, accessKey, credentialScope, signedHeadersStr, signature)
	req.Header.Set("Authorization", authHeader)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(secretKey, dateStr, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStr))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}
