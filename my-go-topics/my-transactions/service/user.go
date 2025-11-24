package user

import (
	"context"
	"errors"
	"fmt"
)

var ErrCreateGroups = errors.New("failed to create group from database")
var ErrSome = errors.New("some error")

func CreateGroup(ctx context.Context, fu func() error) error {

	err := fu()
	if err != nil {
		return Wrap(ErrCreateGroups, err)
	}

	return nil
}

func Wrap(base, ext error) error {
	return fmt.Errorf("%w: %w", base, ext)
}
