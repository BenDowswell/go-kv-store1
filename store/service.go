package store

import (
	"strings"
)

type KVService struct {
	repo KVRepository
}

func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}
	return nil
}

func validateValue(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrEmptyValue
	}
	return nil
}

func NewKVService(repo KVRepository) *KVService {
	return &KVService{
		repo: repo,
	}
}

func (k *KVService) Get(key string) (string, error) {
	err := validateKey(key)
	if err != nil {
		return "", err
	}
	return k.repo.Get(key)
}

func (k *KVService) Set(key, value string) error {
	err := validateKey(key)
	if err != nil {
		return err
	}

	err = validateValue(value)
	if err != nil {
		return err
	}
	return k.repo.Set(key, value)
}

func (k *KVService) Delete(key string) error {
	err := validateKey(key)
	if err != nil {
		return err
	}
	return k.repo.Delete(key)
}
