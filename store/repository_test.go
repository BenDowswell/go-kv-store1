package store

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestMemoryStore_SetGet(t *testing.T) {
	m := NewMemoryStore()
	if err := m.Set("a", "1"); err != nil {
		t.Fatal(err)
	}
	got, err := m.Get("a")
	if err != nil || got != "1" {
		t.Fatalf("got %q, %v", got, err)
	}
	_ = m.Set("a", "2")
	if got, _ := m.Get("a"); got != "2" {
		t.Fatalf("overwrite failed, got %q", got)
	}
}

func TestMemoryStore_GetMissing(t *testing.T) {
	_, err := NewMemoryStore().Get("nope")
	if !errors.Is(err, ErrNoKeyFound) {
		t.Fatalf("got %v", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	m := NewMemoryStore()
	_ = m.Set("a", "1")
	if err := m.Delete("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get("a"); !errors.Is(err, ErrNoKeyFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := m.Delete("a"); err != nil {
		t.Fatalf("deleting missing key should be a no-op, got %v", err)
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	m := NewMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			k := fmt.Sprintf("k%d", i%5)
			_ = m.Set(k, "v")
			_, _ = m.Get(k)
			_ = m.Delete(k)
		}(i)
	}
	wg.Wait()
}
