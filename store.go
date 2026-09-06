package main

import (
	"errors"
	"sync"
)

type Flag struct {
	Key            string `json:"key"`
	Description    string `json:"description"`
	Enabled        bool   `json:"enabled"`
	RolloutPercent int    `json:"rollout_percent"`
}

type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

var errFlagExists = errors.New("flag already exists")

func NewStore() *Store {
	return &Store{flags: make(map[string]Flag)}
}

func (s *Store) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[f.Key]; exists {
		return errFlagExists
	}
	s.flags[f.Key] = f
	return nil
}

func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	flags := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		flags = append(flags, f)
	}
	return flags
}

func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

func (s *Store) Update(key string, f Flag) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return false
	}
	s.flags[key] = f
	return true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return false
	}
	delete(s.flags, key)
	return true
}
