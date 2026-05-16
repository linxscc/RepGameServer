package service

import (
	"encoding/json"
	"fmt"
)

type IdempotencyStore struct{}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{}
}

// CheckAndSet returns true if the key already exists (duplicate), false if this is a new request.
// For new requests, it stores the key with a placeholder response.
func (s *IdempotencyStore) CheckAndSet(key string) (isDuplicate bool, err error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}
	
	// Clean up old keys first (best-effort)
	_, _ = db.Exec(`DELETE FROM voyara_idempotency_keys WHERE created_at < NOW() - INTERVAL 24 HOUR`)

	res, err := db.Exec(`INSERT IGNORE INTO voyara_idempotency_keys (idempotent_key, response) VALUES (?, '{}')`, key)
	if err != nil {
		return false, fmt.Errorf("idempotency check: %v", err)
	}
	rows, _ := res.RowsAffected()
	// If no rows were inserted, the key already existed (duplicate)
	return rows == 0, nil
}

// MarkDone updates the response for a completed idempotent operation
func (s *IdempotencyStore) MarkDone(key string, response interface{}) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
		respJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("idempotency marshal response: %v", err)
	}
	_, err = db.Exec(`UPDATE voyara_idempotency_keys SET response = ? WHERE idempotent_key = ?`, string(respJSON), key)
	return err
}
