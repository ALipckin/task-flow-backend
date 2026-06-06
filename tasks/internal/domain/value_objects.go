package domain

import "unicode/utf8"

const (
	maxTitleLength       = 255
	maxDescriptionLength = 10000
)

func validateTitle(title string) error {
	if title == "" {
		return ErrEmptyTitle
	}
	if utf8.RuneCountInString(title) > maxTitleLength {
		return ErrTitleTooLong
	}
	return nil
}

func validateDescription(description string) error {
	if utf8.RuneCountInString(description) > maxDescriptionLength {
		return ErrDescriptionTooLong
	}
	return nil
}
