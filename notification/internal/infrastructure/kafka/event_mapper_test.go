package kafka

import (
	"testing"

	"notification/internal/domain"
)

func TestDecodeTaskEvent_UsesTaskID(t *testing.T) {
	payload := []byte(`{"event":"TaskUpdated","task_id":5,"performer_id":20,"creator_id":10,"observers_ids":[30]}`)

	event, err := decodeTaskEvent(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.TaskID != 5 {
		t.Fatalf("expected task id 5, got %d", event.TaskID)
	}
	if event.Event != domain.EventTaskUpdated {
		t.Fatalf("expected TaskUpdated, got %s", event.Event)
	}
}

func TestDecodeTaskEvent_FallsBackToID(t *testing.T) {
	payload := []byte(`{"event":"TaskCreated","id":9,"title":"Fix bug","performer_id":20,"creator_id":10}`)

	event, err := decodeTaskEvent(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.TaskID != 9 {
		t.Fatalf("expected task id 9 from id fallback, got %d", event.TaskID)
	}
	if event.Title != "Fix bug" {
		t.Fatalf("expected title, got %s", event.Title)
	}
}
