package use_cases

import (
	"context"
	"errors"
	"testing"
	"time"

	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

func TestCreateTask_PersistsAndPublishesDomainEvent(t *testing.T) {
	repo := newMemRepo()
	cache := newMemCache()
	producer := &memProducer{}
	uc := NewCreateTask(repo, cache, producer, &memAllocator{})

	task, err := uc.Execute(context.Background(), commands.CreateTaskCommand{
		Title:       "Fix bug",
		Description: "Details",
		CreatorID:   10,
		PerformerID: 20,
		ObserverIDs: []uint{30},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID == 0 {
		t.Fatal("expected allocated id")
	}
	if len(producer.created) != 1 {
		t.Fatalf("expected created event, got %d", len(producer.created))
	}
	if _, err := cache.GetTask(context.Background(), task.ID); err != nil {
		t.Fatalf("expected task in cache: %v", err)
	}
}

func TestCreateTask_RejectsInvalidDomain(t *testing.T) {
	uc := NewCreateTask(newMemRepo(), newMemCache(), &memProducer{}, &memAllocator{})

	_, err := uc.Execute(context.Background(), commands.CreateTaskCommand{
		Title:       "",
		CreatorID:   10,
		PerformerID: 20,
	})
	if !errors.Is(err, domain.ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestUpdateTask_AppliesDomainRulesAndPublishes(t *testing.T) {
	repo := newMemRepo()
	existing := domain.ReconstituteTask(1, "Old", "", 10, 20, nil, domain.TaskStatusPending, time.Now(), time.Now(), nil)
	_ = repo.Save(context.Background(), existing)

	cache := newMemCache()
	_ = cache.SetTask(context.Background(), existing)
	producer := &memProducer{}
	uc := NewUpdateTask(repo, cache, producer)

	task, err := uc.Execute(context.Background(), commands.UpdateTaskCommand{
		ID:          1,
		Title:       "New title",
		Description: "New desc",
		Status:      string(domain.TaskStatusInProgress),
		PerformerID: 21,
		CreatorID:   10,
		ObserverIDs: []uint64{30},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "New title" {
		t.Fatalf("expected updated title, got %s", task.Title)
	}
	if len(producer.updated) != 1 {
		t.Fatalf("expected updated event, got %d", len(producer.updated))
	}
	if _, err := cache.GetTask(context.Background(), 1); !errors.Is(err, out.ErrNotFound) {
		t.Fatal("expected cache to be invalidated")
	}
}

func TestDeleteTask_UsesDomainSoftDelete(t *testing.T) {
	repo := newMemRepo()
	existing := domain.ReconstituteTask(1, "Title", "", 10, 20, nil, domain.TaskStatusPending, time.Now(), time.Now(), nil)
	_ = repo.Save(context.Background(), existing)

	producer := &memProducer{}
	index := newMemShardIndex()
	_ = index.Set(context.Background(), 1, 0)
	uc := NewDeleteTask(repo, newMemCache(), index, producer)

	ok, err := uc.Execute(context.Background(), commands.DeleteTaskCommand{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected delete to succeed")
	}
	if !repo.tasks[1].IsDeleted() {
		t.Fatal("expected aggregate to be marked deleted")
	}
	if len(producer.deleted) != 1 {
		t.Fatalf("expected deleted event, got %d", len(producer.deleted))
	}
	if _, err := repo.GetByID(context.Background(), 1); !errors.Is(err, out.ErrNotFound) {
		t.Fatal("expected deleted task to be hidden from get")
	}
}

func TestGetTask_ReturnsCachedTaskWithoutRepository(t *testing.T) {
	repo := newMemRepo()
	cache := newMemCache()
	cached := domain.ReconstituteTask(7, "Cached", "", 10, 20, nil, domain.TaskStatusPending, time.Now(), time.Now(), nil)
	_ = cache.SetTask(context.Background(), cached)

	uc := NewGetTask(repo, cache)
	task, err := uc.Execute(context.Background(), queries.GetTaskQuery{ID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Cached" {
		t.Fatalf("expected cached task, got %s", task.Title)
	}
}
