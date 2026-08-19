package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewTask_ValidInput(t *testing.T) {
	task, err := NewTask(1, "Fix bug", "Details", 10, 20, []uint{30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Status != TaskStatusPending {
		t.Fatalf("expected pending status, got %s", task.Status)
	}
	if len(task.Observers) != 1 || task.Observers[0].UserID != 30 {
		t.Fatalf("unexpected observers: %+v", task.Observers)
	}

	events := task.PullEvents()
	if len(events) != 1 {
		t.Fatalf("expected one domain event, got %d", len(events))
	}
	if events[0].Name() != "TaskCreated" {
		t.Fatalf("expected TaskCreated event, got %s", events[0].Name())
	}
}

func TestNewTask_RejectsEmptyTitle(t *testing.T) {
	_, err := NewTask(1, "", "Details", 10, 20, nil)
	if !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestNewTask_RejectsObserverAsPerformer(t *testing.T) {
	_, err := NewTask(1, "Title", "Details", 10, 20, []uint{20})
	if !errors.Is(err, ErrObserverIsPerformer) {
		t.Fatalf("expected ErrObserverIsPerformer, got %v", err)
	}
}

func TestTask_ChangeStatus_ValidTransitions(t *testing.T) {
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusPending, now(), now(), nil)

	if err := task.changeStatus(string(TaskStatusInProgress)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != TaskStatusInProgress {
		t.Fatalf("expected in_progress, got %s", task.Status)
	}
}

func TestTask_ChangeStatus_InvalidTransition(t *testing.T) {
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusDone, now(), now(), nil)

	err := task.changeStatus(string(TaskStatusInProgress))
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestTask_SetObservers_Deduplicates(t *testing.T) {
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusPending, now(), now(), nil)

	err := task.setObservers([]uint{30, 30})
	if !errors.Is(err, ErrDuplicateObserver) {
		t.Fatalf("expected ErrDuplicateObserver, got %v", err)
	}
}

func TestTask_ApplyUpdate_RecordsEvent(t *testing.T) {
	task := ReconstituteTask(1, "Title", "Old", 10, 20, nil, TaskStatusPending, now(), now(), nil)

	if err := task.ApplyUpdate("New title", "New desc", string(TaskStatusInProgress), 21, 11, []uint{30}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.PerformerChanged() {
		t.Fatal("expected performer change to be tracked")
	}
	if task.PreviousPerformerID() != 20 {
		t.Fatalf("expected previous performer 20, got %d", task.PreviousPerformerID())
	}

	events := task.PullEvents()
	if len(events) != 1 || events[0].Name() != "TaskUpdated" {
		t.Fatalf("expected TaskUpdated event, got %+v", events)
	}
}

func TestTask_SetObservers_RejectsTooMany(t *testing.T) {
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusPending, now(), now(), nil)
	ids := make([]uint, maxObservers+1)
	for i := range ids {
		ids[i] = uint(100 + i)
	}

	err := task.setObservers(ids)
	if !errors.Is(err, ErrTooManyObservers) {
		t.Fatalf("expected ErrTooManyObservers, got %v", err)
	}
}

func TestTask_MarkDeleted_RecordsEvent(t *testing.T) {
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusPending, now(), now(), nil)

	if err := task.MarkDeleted(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.IsDeleted() {
		t.Fatal("expected task to be deleted")
	}

	events := task.PullEvents()
	if len(events) != 1 || events[0].Name() != "TaskDeleted" {
		t.Fatalf("expected TaskDeleted event, got %+v", events)
	}
}

func TestTask_ApplyUpdate_RejectedWhenDeleted(t *testing.T) {
	deletedAt := now()
	task := ReconstituteTask(1, "Title", "", 10, 20, nil, TaskStatusPending, now(), now(), &deletedAt)

	err := task.ApplyUpdate("New", "Desc", string(TaskStatusInProgress), 21, 11, nil)
	if !errors.Is(err, ErrTaskDeleted) {
		t.Fatalf("expected ErrTaskDeleted, got %v", err)
	}
}

func TestParseTaskStatus_AcceptsNewAlias(t *testing.T) {
	status, err := ParseTaskStatus("new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != TaskStatusPending {
		t.Fatalf("expected pending, got %s", status)
	}
}

func now() time.Time {
	return time.Now()
}
