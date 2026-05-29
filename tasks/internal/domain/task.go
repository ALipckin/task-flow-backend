package domain

import "time"

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
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewTask(id uint, title, description string, creatorID, performerID uint) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		PerformerId: performerID,
		CreatorId:   creatorID,
		Status:      "new",
	}
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
