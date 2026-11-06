package service

import (
	"testing"
	"time"
)

func TestSessionStoreIssueResolveExpire(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := NewSessionStore(func() time.Time { return now })

	tok, _, err := store.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	if user, ok := store.Resolve(tok); !ok || user != "alice" {
		t.Fatalf("resolve = %q, %v", user, ok)
	}

	now = now.Add(SessionTTL() + time.Second)
	if _, ok := store.Resolve(tok); ok {
		t.Fatal("expired token should not resolve")
	}
}

func TestSessionStoreReissueRevokesOldToken(t *testing.T) {
	store := NewSessionStore(time.Now)
	first, _, _ := store.Issue("bob")
	second, _, _ := store.Issue("bob")

	if first == second {
		t.Fatal("tokens must differ")
	}
	if _, ok := store.Resolve(first); ok {
		t.Fatal("old token should be revoked")
	}
	if user, ok := store.Resolve(second); !ok || user != "bob" {
		t.Fatal("new token should resolve")
	}
}

func TestValidateCredentialFormat(t *testing.T) {
	cases := []struct {
		user, pass string
		ok         bool
	}{
		{"alice", "secret1", true},
		{"ab", "secret1", false},
		{"alice", "12345", false},
		{" alice", "secret1", false},
		{"玩家一号", "secret1", true},
	}
	for _, c := range cases {
		if err := ValidateCredentialFormat(c.user, c.pass); (err == nil) != c.ok {
			t.Errorf("ValidateCredentialFormat(%q,%q) err=%v want ok=%v", c.user, c.pass, err, c.ok)
		}
	}
}

func TestBcryptHashDetection(t *testing.T) {
	hash, err := HashPassword("secret1")
	if err != nil {
		t.Fatal(err)
	}
	if !isBcryptHash(hash) {
		t.Fatalf("expected bcrypt hash, got %q", hash)
	}
	if isBcryptHash("secret1") {
		t.Fatal("plaintext must not be detected as bcrypt")
	}
}
