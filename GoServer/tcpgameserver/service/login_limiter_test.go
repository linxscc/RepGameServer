package service

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksWithinWindowAndResets(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewRateLimiter(3, time.Minute, func() time.Time { return now })

	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("attempt %d should pass", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4th attempt within the window must be rejected")
	}
	if !l.Allow("5.6.7.8") {
		t.Fatal("other keys are limited independently")
	}

	now = now.Add(time.Minute)
	if !l.Allow("1.2.3.4") {
		t.Fatal("a new window should allow attempts again")
	}
}
