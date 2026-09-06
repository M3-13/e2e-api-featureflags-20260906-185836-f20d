package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "a", Description: "d", Enabled: true, RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if got.Key != "a" || got.Description != "d" || !got.Enabled || got.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", got)
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	s := NewStore()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Create(Flag{Key: "a"}); err == nil {
		t.Fatal("expected duplicate key to error")
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	flags := s.List()
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a", Enabled: false})
	if !s.Update("a", Flag{Key: "a", Enabled: true, RolloutPercent: 10}) {
		t.Fatal("expected update of existing key to succeed")
	}
	got, _ := s.Get("a")
	if !got.Enabled || got.RolloutPercent != 10 {
		t.Fatalf("unexpected flag after update: %+v", got)
	}
	if s.Update("missing", Flag{Key: "missing"}) {
		t.Fatal("expected update of missing key to fail")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a"})
	if !s.Delete("a") {
		t.Fatal("expected delete of existing key to succeed")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("expected flag to be gone after delete")
	}
	if s.Delete("a") {
		t.Fatal("expected delete of missing key to fail")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key-%d", i)
		wg.Add(2)
		go func(k string) {
			defer wg.Done()
			_ = s.Create(Flag{Key: k})
		}(key)
		go func(k string) {
			defer wg.Done()
			s.Update(k, Flag{Key: k, Enabled: true})
			_, _ = s.Get(k)
		}(key)
	}
	wg.Wait()
}

func TestStoreMaxFlags(t *testing.T) {
	s := NewStore()
	for i := 0; i < maxFlags; i++ {
		if err := s.Create(Flag{Key: fmt.Sprintf("k%d", i)}); err != nil {
			t.Fatalf("create %d: unexpected error: %v", i, err)
		}
	}
	if err := s.Create(Flag{Key: "overflow"}); !errors.Is(err, errMaxFlags) {
		t.Fatalf("expected errMaxFlags, got %v", err)
	}
}

func TestUpdateIfExistsUnknownKey(t *testing.T) {
	s := NewStore()
	if s.UpdateIfExists("missing", Flag{Key: "missing"}) {
		t.Fatal("expected UpdateIfExists on unknown key to return false")
	}
}

func TestUpdateIfExistsExistingKey(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a", Enabled: false})
	if !s.UpdateIfExists("a", Flag{Key: "a", Enabled: true}) {
		t.Fatal("expected UpdateIfExists on existing key to succeed")
	}
	got, _ := s.Get("a")
	if !got.Enabled {
		t.Fatalf("unexpected flag after UpdateIfExists: %+v", got)
	}
}
