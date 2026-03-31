package persistence

import "time"

type Task struct {
	// TODO change ids to int64
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

func (t *Task) ObserverIDs() []uint64 {
	if len(t.Observers) == 0 {
		return nil
	}

	observerIds := make([]uint64, len(t.Observers))
	for i, observer := range t.Observers {
		observerIds[i] = uint64(observer.UserId)
	}

	return observerIds
}

func ObserversFromIDs(observerIds []uint64) []Observer {
	observers := make([]Observer, len(observerIds))
	for i, observerId := range observerIds {
		observers[i] = Observer{UserId: uint(observerId)}
	}
	return observers
}
