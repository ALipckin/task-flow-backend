package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type GetTasks struct {
	repo out.Repository
}

func NewGetTasks(repo out.Repository) *GetTasks {
	return &GetTasks{repo: repo}
}

func (uc *GetTasks) Execute(ctx context.Context, cmd queries.GetTasksQuery) ([]domain.Task, error) {
	filter := out.TaskFilter{
		Title:       cmd.Title,
		CreatorID:   cmd.CreatorID,
		PerformerID: cmd.PerformerID,
	}

	return uc.repo.Find(ctx, filter)
}
