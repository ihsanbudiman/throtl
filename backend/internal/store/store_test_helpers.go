package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"

	_ "modernc.org/sqlite"
)

// NewTestStore creates an in-memory SQLite Store with all tables migrated.
// It is safe for concurrent use within a single test. Call s.Close() in cleanup.
func NewTestStore() (*Store, error) {
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open test db: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate test db: %w", err)
	}
	return s, nil
}

// MustTestStore calls NewTestStore and panics on error.
// Useful for table-driven tests where you want a fresh DB per case.
func MustTestStore() *Store {
	s, err := NewTestStore()
	if err != nil {
		panic(err)
	}
	return s
}

// InsertTestProvider inserts a minimal provider fixture and returns it.
func InsertTestProvider(s *Store, id string) *model.Provider {
	p := &model.Provider{
		ID:        id,
		Name:      "test-provider",
		Type:      "openai",
		BaseURL:   "https://api.test.example.com/v1",
		APIKey:    "test-provider-key",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.CreateProvider(p); err != nil {
		panic(fmt.Sprintf("InsertTestProvider: %v", err))
	}
	return p
}

// InsertTestAPIKey inserts a minimal API key fixture and returns it.
func InsertTestAPIKey(s *Store, id, shareKey string) *model.APIKey {
	k := &model.APIKey{
		ID:            id,
		Name:          "test-key",
		Key:           shareKey,
		LimitWindow:   100,
		LimitDaily:    500,
		LimitWindowHrs: 5,
		AllowedModels: "",
		Active:        true,
		CreatedAt:     time.Now().UTC().Truncate(time.Second),
	}
	if err := s.CreateAPIKey(k); err != nil {
		panic(fmt.Sprintf("InsertTestAPIKey: %v", err))
	}
	return k
}
