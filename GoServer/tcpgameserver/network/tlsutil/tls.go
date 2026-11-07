// Package tlsutil 负责加载（或首次生成）服务器 TLS 证书。
package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options 证书来源配置。
type Options struct {
	CertFile string
	KeyFile  string
	// Hosts 自签名证书的 SAN（IP 或域名），仅在需要生成证书时使用。
	Hosts []string
}

// ServerConfig 返回服务器 TLS 配置以及证书 SHA-256 指纹（供客户端固定证书）。
// 证书文件不存在时自动生成一张自签名证书并写入磁盘。
func ServerConfig(opts Options) (*tls.Config, string, error) {
	if _, err := os.Stat(opts.CertFile); errors.Is(err, os.ErrNotExist) {
		if err := generateSelfSigned(opts); err != nil {
			return nil, "", fmt.Errorf("generate self-signed certificate: %w", err)
		}
	}

	cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
	if err != nil {
		return nil, "", fmt.Errorf("load certificate: %w", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	return cfg, Fingerprint(cert.Certificate[0]), nil
}

// Fingerprint 证书 DER 的 SHA-256，十六进制小写、无分隔符。
func Fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}

func generateSelfSigned(opts Options) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "RepGame Server"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range opts.Hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(opts.CertFile), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opts.KeyFile), 0o700); err != nil {
		return err
	}
	if err := writePEM(opts.KeyFile, "EC PRIVATE KEY", keyDER, 0o600); err != nil {
		return err
	}
	return writePEM(opts.CertFile, "CERTIFICATE", der, 0o644)
}

func writePEM(path, blockType string, der []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: der})
}
