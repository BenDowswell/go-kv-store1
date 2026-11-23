package store

import (
	"errors"
	"testing"
)

type fakeRepo struct {
	calls int
	err   error
}

func (f *fakeRepo) Get(string) (string, error) { f.calls++; return "v", f.err }
func (f *fakeRepo) Set(string, string) error   { f.calls++; return f.err }
func (f *fakeRepo) Delete(string) error        { f.calls++; return f.err }

func TestService_Validation(t *testing.T) {
	tests := []struct {
		name       string
		key, value string
		wantSet    error
	}{
		{"ok", "k", "v", nil},
		{"empty key", "", "v", ErrEmptyKey},
		{"blank key", "   ", "v", ErrEmptyKey},
		{"empty value", "k", "", ErrEmptyValue},
		{"blank value", "k", " \t", ErrEmptyValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := NewKVService(repo)
			err := svc.Set(tt.key, tt.value)
			if !errors.Is(err, tt.wantSet) {
				t.Fatalf("Set err = %v, want %v", err, tt.wantSet)
			}
			if tt.wantSet != nil && repo.calls != 0 {
				t.Fatal("repo called despite validation failure")
			}
		})
	}
}

func TestService_GetDeleteEmptyKey(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewKVService(repo)
	if _, err := svc.Get(" "); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Get: %v", err)
	}
	if err := svc.Delete(""); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Delete: %v", err)
	}
	if repo.calls != 0 {
		t.Fatal("repo called")
	}
}

func TestService_PassesRepoErrors(t *testing.T) {
	svc := NewKVService(&fakeRepo{err: ErrSavingKeyValue})
	if err := svc.Set("k", "v"); !errors.Is(err, ErrSavingKeyValue) {
		t.Fatalf("got %v", err)
	}
	if _, err := svc.Get("k"); !errors.Is(err, ErrSavingKeyValue) {
		t.Fatalf("got %v", err)
	}
}
