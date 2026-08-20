package initializers

import (
	"errors"
	"testing"
	"time"
)

func TestRetryKafkaConnect_SucceedsAfterTransientFailures(t *testing.T) {
	t.Parallel()

	calls := 0
	client, err := retryKafkaConnect("producer", 5, 0, func() (string, error) {
		calls++
		if calls < 3 {
			return "", errors.New("connection refused")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if client != "ok" {
		t.Fatalf("expected ok client, got %q", client)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestRetryKafkaConnect_FailsAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	want := errors.New("connection refused")
	calls := 0
	_, err := retryKafkaConnect("consumer", 3, 0, func() (string, error) {
		calls++
		return "", want
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestRetryKafkaConnect_SucceedsOnFirstAttempt(t *testing.T) {
	t.Parallel()

	calls := 0
	start := time.Now()
	_, err := retryKafkaConnect("producer", 5, time.Second, func() (string, error) {
		calls++
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 attempt, got %d", calls)
	}
	if time.Since(start) >= 500*time.Millisecond {
		t.Fatal("did not expect retry delay on first success")
	}
}
