package kafke

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestConnectWithRetry_SucceedsAfterTransientFailures(t *testing.T) {
	t.Parallel()

	calls := 0
	factory := func([]string, *sarama.Config) (sarama.SyncProducer, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("connection refused")
		}
		return nil, nil
	}

	producer, err := connectWithRetry([]string{"kafka:9092"}, sarama.NewConfig(), factory, 5, 0)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if producer != nil {
		t.Fatal("expected nil mock producer")
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestConnectWithRetry_FailsAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	want := errors.New("connection refused")
	calls := 0
	factory := func([]string, *sarama.Config) (sarama.SyncProducer, error) {
		calls++
		return nil, want
	}

	_, err := connectWithRetry([]string{"kafka:9092"}, sarama.NewConfig(), factory, 3, 0)
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

func TestConnectWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	t.Parallel()

	calls := 0
	factory := func([]string, *sarama.Config) (sarama.SyncProducer, error) {
		calls++
		return nil, nil
	}

	start := time.Now()
	_, err := connectWithRetry([]string{"kafka:9092"}, sarama.NewConfig(), factory, 5, time.Second)
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
