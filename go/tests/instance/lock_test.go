package instance_test

import (
	"errors"
	"path/filepath"
	"testing"

	"herdr-theme-picker/internal/instance"
)

func TestExclusiveLockAndRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.lock")
	first, err := instance.Acquire(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { first.Close() })
	second, err := instance.Acquire(path)
	if second != nil {
		second.Close()
		t.Fatal("a second daemon acquired the same lock")
	}
	if !errors.Is(err, instance.ErrLocked) {
		t.Fatalf("expected lock conflict: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	third, err := instance.Acquire(path)
	if err != nil {
		t.Fatalf("lock was not released on close: %v", err)
	}
	third.Close()
}
