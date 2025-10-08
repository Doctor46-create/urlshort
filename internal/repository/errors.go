package repository

import "fmt"

type URLConflictError struct {
	ExistingShortURL string
}

func (e *URLConflictError) Error() string {
	return fmt.Sprintf("URL already exists: %s", e.ExistingShortURL)
}

func NewURLConflictError(existingShortURL string) error {
	return &URLConflictError{ExistingShortURL: existingShortURL}
}

func IsURLConflictError(err error) (string, bool) {
	if conflictErr, ok := err.(*URLConflictError); ok {
		return conflictErr.ExistingShortURL, true
	}
	return "", false
}

