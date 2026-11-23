package store

import (
	"errors"
)

var (
	ErrNoKeyFound     = errors.New("no key found")
	ErrSavingKeyValue = errors.New("error when saving")
	ErrDeletingKey    = errors.New("error when deleting")
	ErrEmptyKey       = errors.New("key must not be empty")
	ErrEmptyValue     = errors.New("value must not be empty")
)
