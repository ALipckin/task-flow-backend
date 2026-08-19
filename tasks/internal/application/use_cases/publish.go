package use_cases

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

func publishTaskEvents(ctx context.Context, producer out.EventProducer, task *domain.Task) error {
	if producer == nil || task == nil {
		return nil
	}

	for _, event := range task.PullEvents() {
		var err error
		switch event.(type) {
		case domain.TaskCreated:
			err = producer.PublishCreated(ctx, *task)
		case domain.TaskUpdated:
			err = producer.PublishUpdated(ctx, *task)
		case domain.TaskDeleted:
			err = producer.PublishDeleted(ctx, *task)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
