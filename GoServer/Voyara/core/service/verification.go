package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"crypto/rand"
	"time"
)

const (
	codeLength    = 6
	codeExpiry    = 5 * time.Minute
	codeCooldown  = 60 * time.Second
)

func GenerateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < codeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate code: %v", err)
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}

func SaveVerificationCode(email, purpose string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Check cooldown: don't allow resend within 60 seconds
	var lastCreated time.Time
	err = db.QueryRow(`
		SELECT created_at FROM voyara_verification_codes
		WHERE email = ? AND purpose = ? AND used = 0
		ORDER BY created_at DESC LIMIT 1`, email, purpose).Scan(&lastCreated)
	if err == nil && time.Since(lastCreated) < codeCooldown {
		remaining := int(codeCooldown.Seconds() - time.Since(lastCreated).Seconds())
		return "", fmt.Errorf("please wait %d seconds before requesting a new code", remaining)
	}

	code, err := GenerateVerificationCode()
	if err != nil {
		return "", err
	}

	_, err = db.Exec(`
		INSERT INTO voyara_verification_codes (email, code, purpose, expires_at)
		VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL 5 MINUTE))`, email, code, purpose)
	if err != nil {
		return "", fmt.Errorf("save code: %v", err)
	}

	return code, nil
}

func VerifyCode(email, code, purpose string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var id int64
	err = db.QueryRow(`
		SELECT id FROM voyara_verification_codes
		WHERE email = ? AND code = ? AND purpose = ? AND used = 0 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1`, email, code, purpose).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("verify code: %v", err)
	}

	// Mark as used
	_, err = db.Exec(`UPDATE voyara_verification_codes SET used = 1 WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("mark code used: %v", err)
	}

	return true, nil
}

func MarkEmailVerified(email string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`UPDATE voyara_users SET email_verified_at = NOW() WHERE email = ?`, email)
	return err
}
