package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrAlreadyCompleted = errors.New("already completed")
var ErrInvalidTitle = errors.New("title cannot be empty")