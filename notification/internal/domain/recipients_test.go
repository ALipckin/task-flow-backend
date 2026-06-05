package domain

import "testing"

func TestNotificationRecipients_IncludesStakeholders(t *testing.T) {
	event := TaskEvent{
		ObserversIDs: []int{30, 40},
		PerformerID:  20,
		CreatorID:    10,
	}

	recipients := NotificationRecipients(event)
	if len(recipients) != 4 {
		t.Fatalf("expected 4 recipients, got %d: %v", len(recipients), recipients)
	}
}

func TestNotificationRecipients_Deduplicates(t *testing.T) {
	event := TaskEvent{
		ObserversIDs: []int{10, 20},
		PerformerID:  20,
		CreatorID:    10,
	}

	recipients := NotificationRecipients(event)
	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d: %v", len(recipients), recipients)
	}
}
