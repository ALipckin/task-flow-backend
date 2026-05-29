package persistence

import "tasks/internal/domain"

func TaskToDomain(t Task) domain.Task {
	return domain.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		PerformerId: t.PerformerId,
		CreatorId:   t.CreatorId,
		Observers:   ObserversToDomain(t.Observers),
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DeletedAt:   t.DeletedAt,
	}
}

func TaskFromDomain(t domain.Task) Task {
	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		PerformerId: t.PerformerId,
		CreatorId:   t.CreatorId,
		Observers:   ObserversFromDomain(t.Observers),
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DeletedAt:   t.DeletedAt,
	}
}

func ObserversToDomain(observers []Observer) []domain.Observer {
	if len(observers) == 0 {
		return nil
	}

	out := make([]domain.Observer, len(observers))
	for i, o := range observers {
		out[i] = domain.Observer{UserID: o.UserId}
	}
	return out
}

func ObserversFromDomain(observers []domain.Observer) []Observer {
	if len(observers) == 0 {
		return nil
	}

	out := make([]Observer, len(observers))
	for i, o := range observers {
		out[i] = Observer{UserId: o.UserID}
	}
	return out
}

func ObserversFromUserIDs(ids []uint) []Observer {
	if len(ids) == 0 {
		return nil
	}

	observers := make([]Observer, len(ids))
	for i, id := range ids {
		observers[i] = Observer{UserId: id}
	}
	return observers
}
