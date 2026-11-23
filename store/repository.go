package store

import "sync"

type KVRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[string]string),
	}
}

func (m *MemoryStore) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.store[key]
	if !ok {
		return "", ErrNoKeyFound
	}
	return value, nil
}

func (m *MemoryStore) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = value
	return nil
}

func (m *MemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}
