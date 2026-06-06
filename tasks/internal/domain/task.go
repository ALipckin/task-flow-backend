package domain

import "time"

const maxObservers = 50

type Observer struct {
	UserID uint
}

type Task struct {
	ID          uint
	Title       string
	Description string
	PerformerId uint
	CreatorId   uint
	Observers   []Observer
	Status      TaskStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	events              []DomainEvent
	previousPerformerID uint
}

func NewTask(
	id uint,
	title, description string,
	creatorID, performerID uint,
	observerIDs []uint,
) (Task, error) {
	if creatorID == 0 {
		return Task{}, ErrInvalidCreator
	}
	if performerID == 0 {
		return Task{}, ErrInvalidPerformer
	}
	if err := validateTitle(title); err != nil {
		return Task{}, err
	}
	if err := validateDescription(description); err != nil {
		return Task{}, err
	}

	now := time.Now()
	task := Task{
		ID:          id,
		Title:       title,
		Description: description,
		PerformerId: performerID,
		CreatorId:   creatorID,
		Status:      TaskStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := task.setObservers(observerIDs); err != nil {
		return Task{}, err
	}

	task.recordEvent(TaskCreated{
		TaskID:      task.ID,
		Title:       task.Title,
		CreatorID:   task.CreatorId,
		PerformerID: task.PerformerId,
		OccurredAt:  now,
	})

	return task, nil
}

func ReconstituteTask(
	id uint,
	title, description string,
	creatorID, performerID uint,
	observers []Observer,
	status TaskStatus,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		PerformerId: performerID,
		CreatorId:   creatorID,
		Observers:   observers,
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	}
}

func (t *Task) ApplyUpdate(
	title, description, status string,
	performerID, creatorID uint,
	observerIDs []uint,
) error {
	if t.IsDeleted() {
		return ErrTaskDeleted
	}
	if err := t.setTitle(title); err != nil {
		return err
	}
	if err := t.setDescription(description); err != nil {
		return err
	}
	if err := t.changeStatus(status); err != nil {
		return err
	}
	if err := t.assignPerformer(performerID); err != nil {
		return err
	}
	if err := t.changeCreator(creatorID); err != nil {
		return err
	}
	if err := t.setObservers(observerIDs); err != nil {
		return err
	}

	t.touch()
	t.recordEvent(TaskUpdated{TaskID: t.ID, OccurredAt: t.UpdatedAt})
	return nil
}

func (t *Task) MarkDeleted() error {
	if t.IsDeleted() {
		return ErrTaskDeleted
	}

	now := time.Now()
	t.DeletedAt = &now
	t.touch()
	t.recordEvent(TaskDeleted{TaskID: t.ID, OccurredAt: now})
	return nil
}

func (t Task) IsDeleted() bool {
	return t.DeletedAt != nil
}

func (t Task) PerformerChanged() bool {
	return t.previousPerformerID != 0 && t.previousPerformerID != t.PerformerId
}

func (t Task) PreviousPerformerID() uint {
	return t.previousPerformerID
}

func (t *Task) PullEvents() []DomainEvent {
	events := t.events
	t.events = nil
	return events
}

func (t Task) ObserverUserIDs() []uint64 {
	if len(t.Observers) == 0 {
		return nil
	}

	ids := make([]uint64, len(t.Observers))
	for i, o := range t.Observers {
		ids[i] = uint64(o.UserID)
	}
	return ids
}

func ObserversFromUserIDs(ids []uint64) []Observer {
	if len(ids) == 0 {
		return nil
	}

	observers := make([]Observer, len(ids))
	for i, id := range ids {
		observers[i] = Observer{UserID: uint(id)}
	}
	return observers
}

func (t *Task) setTitle(title string) error {
	if err := validateTitle(title); err != nil {
		return err
	}
	t.Title = title
	return nil
}

func (t *Task) setDescription(description string) error {
	if err := validateDescription(description); err != nil {
		return err
	}
	t.Description = description
	return nil
}

func (t *Task) changeStatus(raw string) error {
	if raw == "" {
		return nil
	}

	status, err := ParseTaskStatus(raw)
	if err != nil {
		return err
	}
	if !t.Status.CanTransitionTo(status) {
		return ErrInvalidStatusTransition
	}

	t.Status = status
	return nil
}

func (t *Task) assignPerformer(performerID uint) error {
	if performerID == 0 {
		return ErrInvalidPerformer
	}
	if t.PerformerId != performerID {
		t.previousPerformerID = t.PerformerId
		t.PerformerId = performerID
	}
	return nil
}

func (t *Task) changeCreator(creatorID uint) error {
	if creatorID == 0 {
		return ErrInvalidCreator
	}
	t.CreatorId = creatorID
	return nil
}

func (t *Task) setObservers(observerIDs []uint) error {
	if len(observerIDs) == 0 {
		t.Observers = nil
		return nil
	}
	if len(observerIDs) > maxObservers {
		return ErrDuplicateObserver
	}

	seen := make(map[uint]struct{}, len(observerIDs))
	observers := make([]Observer, 0, len(observerIDs))
	for _, id := range observerIDs {
		if id == 0 {
			continue
		}
		if id == t.PerformerId {
			return ErrObserverIsPerformer
		}
		if _, exists := seen[id]; exists {
			return ErrDuplicateObserver
		}
		seen[id] = struct{}{}
		observers = append(observers, Observer{UserID: id})
	}

	t.Observers = observers
	return nil
}

func (t *Task) touch() {
	t.UpdatedAt = time.Now()
}

func (t *Task) recordEvent(event DomainEvent) {
	t.events = append(t.events, event)
}
