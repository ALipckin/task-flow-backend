package use_cases

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type memRepo struct {
	tasks map[uint]domain.Task
}

func newMemRepo() *memRepo {
	return &memRepo{tasks: make(map[uint]domain.Task)}
}

func (r *memRepo) Save(_ context.Context, task domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *memRepo) Find(_ context.Context, _ out.TaskFilter) ([]domain.Task, error) {
	result := make([]domain.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if !task.IsDeleted() {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *memRepo) Delete(_ context.Context, task domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *memRepo) GetByID(_ context.Context, taskID uint) (*domain.Task, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.IsDeleted() {
		return nil, out.ErrNotFound
	}
	cp := task
	return &cp, nil
}

func (r *memRepo) Update(_ context.Context, task domain.Task) error {
	if _, ok := r.tasks[task.ID]; !ok {
		return out.ErrNotFound
	}
	r.tasks[task.ID] = task
	return nil
}

type memCache struct {
	tasks map[uint]domain.Task
}

func newMemCache() *memCache {
	return &memCache{tasks: make(map[uint]domain.Task)}
}

func (c *memCache) SetTask(_ context.Context, task domain.Task) error {
	c.tasks[task.ID] = task
	return nil
}

func (c *memCache) GetTask(_ context.Context, taskID uint) (domain.Task, error) {
	task, ok := c.tasks[taskID]
	if !ok {
		return domain.Task{}, out.ErrNotFound
	}
	return task, nil
}

func (c *memCache) DeleteTask(_ context.Context, taskID uint) error {
	delete(c.tasks, taskID)
	return nil
}

type memProducer struct {
	created []domain.Task
	updated []domain.Task
	deleted []domain.Task
}

func (p *memProducer) PublishCreated(_ context.Context, task domain.Task) error {
	p.created = append(p.created, task)
	return nil
}

func (p *memProducer) PublishDeleted(_ context.Context, task domain.Task) error {
	p.deleted = append(p.deleted, task)
	return nil
}

func (p *memProducer) PublishUpdated(_ context.Context, task domain.Task) error {
	p.updated = append(p.updated, task)
	return nil
}

type memAllocator struct {
	next uint
}

func (a *memAllocator) NextID(_ context.Context) (uint, error) {
	a.next++
	return a.next, nil
}

type memShardIndex struct {
	ids map[uint]int
}

func newMemShardIndex() *memShardIndex {
	return &memShardIndex{ids: make(map[uint]int)}
}

func (i *memShardIndex) Get(_ context.Context, taskID uint) (int, error) {
	idx, ok := i.ids[taskID]
	if !ok {
		return 0, out.ErrNotFound
	}
	return idx, nil
}

func (i *memShardIndex) Set(_ context.Context, taskID uint, shardIndex int) error {
	i.ids[taskID] = shardIndex
	return nil
}

func (i *memShardIndex) Delete(_ context.Context, taskID uint) error {
	delete(i.ids, taskID)
	return nil
}
