package domain

import "errors"

var (
	ErrEmptyTitle              = errors.New("task title must not be empty")
	ErrTitleTooLong            = errors.New("task title must not exceed 255 characters")
	ErrDescriptionTooLong      = errors.New("task description must not exceed 10000 characters")
	ErrInvalidStatus           = errors.New("invalid task status")
	ErrInvalidStatusTransition = errors.New("invalid task status transition")
	ErrInvalidPerformer        = errors.New("performer id must be greater than zero")
	ErrInvalidCreator          = errors.New("creator id must be greater than zero")
	ErrObserverIsPerformer     = errors.New("observer cannot be the task performer")
	ErrDuplicateObserver       = errors.New("duplicate observer id")
	ErrTaskDeleted             = errors.New("task is deleted")
)
