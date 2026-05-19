package store

import (
	"database/sql"
	"testing"
)

func TestNewTestStore(t *testing.T) {
	s, err := NewTestStore()
	if err != nil {
		t.Fatalf("NewTestStore() error: %v", err)
	}
	defer s.Close()

	wantTables := map[string]bool{
		"api_keys":        false,
		"usage_logs":      false,
		"providers":       false,
		"users":           false,
		"model_overrides": false,
	}

	rows, err := s.db.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	defer rows.Close()

	var found int
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		if _, ok := wantTables[name]; ok {
			wantTables[name] = true
			found++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	if found != len(wantTables) {
		for tbl, seen := range wantTables {
			if !seen {
				t.Errorf("missing table: %s", tbl)
			}
		}
		t.Fatalf("expected %d tables, found %d", len(wantTables), found)
	}
}

func TestMustTestStore(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&count)
	if err != nil {
		t.Fatalf("query providers: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 providers, got %d", count)
	}
}

func TestInsertTestFixtures(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	p := InsertTestProvider(s, "prov-test-1")
	if p.ID != "prov-test-1" {
		t.Errorf("provider id = %q, want %q", p.ID, "prov-test-1")
	}

	got, err := s.GetProvider("prov-test-1")
	if err != nil {
		t.Fatalf("GetProvider: %v", err)
	}
	if got == nil {
		t.Fatal("GetProvider returned nil")
	}
	if got.Name != "test-provider" {
		t.Errorf("provider name = %q, want %q", got.Name, "test-provider")
	}

	k := InsertTestAPIKey(s, "key-test-1", "sk-share-test-1")
	if k.ID != "key-test-1" {
		t.Errorf("key id = %q, want %q", k.ID, "key-test-1")
	}

	gotKey, err := s.GetAPIKeyByShareKey("sk-share-test-1")
	if err != nil {
		t.Fatalf("GetAPIKeyByShareKey: %v", err)
	}
	if gotKey == nil {
		t.Fatal("GetAPIKeyByShareKey returned nil")
	}
	if gotKey.Name != "test-key" {
		t.Errorf("key name = %q, want %q", gotKey.Name, "test-key")
	}
}

func TestNewTestStoreIsolation(t *testing.T) {
	s1 := MustTestStore()
	defer s1.Close()

	InsertTestProvider(s1, "prov-isolated")

	s2 := MustTestStore()
	defer s2.Close()

	got, err := s2.GetProvider("prov-isolated")
	if err != nil {
		t.Fatalf("GetProvider in s2: %v", err)
	}
	if got != nil {
		t.Error("s2 should not see data from s1 (in-memory DBs are isolated)")
	}

	var count int
	if err := s1.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&count); err != nil && err != sql.ErrNoRows {
		t.Fatalf("count providers in s1: %v", err)
	}
	if count != 1 {
		t.Errorf("s1 provider count = %d, want 1", count)
	}
}
