package store

import (
	"testing"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"
)

func TestCreateAPIKeyWithTokenLimits(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := &model.APIKey{
		ID:                  "key-token-1",
		Name:                "token-limited-key",
		Key:                 "sk-share-token-1",
		LimitDaily:          500,
		LimitTokensInDaily:  50000,
		LimitTokensOutDaily: 100000,
		AllowedModels:       "",
		Active:              true,
		CreatedAt:           time.Now().UTC().Truncate(time.Second),
	}
	if err := s.CreateAPIKey(k); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	got, err := s.GetAPIKeyByID("key-token-1")
	if err != nil {
		t.Fatalf("GetAPIKeyByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetAPIKeyByID returned nil")
	}
	if got.LimitTokensInDaily != 50000 {
		t.Errorf("LimitTokensInDaily = %d, want 50000", got.LimitTokensInDaily)
	}
	if got.LimitTokensOutDaily != 100000 {
		t.Errorf("LimitTokensOutDaily = %d, want 100000", got.LimitTokensOutDaily)
	}
	if got.TokensInDailyCount != 0 {
		t.Errorf("TokensInDailyCount = %d, want 0", got.TokensInDailyCount)
	}
	if got.TokensOutDailyCount != 0 {
		t.Errorf("TokensOutDailyCount = %d, want 0", got.TokensOutDailyCount)
	}
}

func TestTokenLimitsDefaultToZero(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := &model.APIKey{
		ID:            "key-default-1",
		Name:          "no-token-limits",
		Key:           "sk-share-default-1",
		LimitDaily:    0,
		AllowedModels: "",
		Active:        true,
		CreatedAt:     time.Now().UTC().Truncate(time.Second),
	}
	if err := s.CreateAPIKey(k); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	got, err := s.GetAPIKeyByID("key-default-1")
	if err != nil {
		t.Fatalf("GetAPIKeyByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetAPIKeyByID returned nil")
	}
	if got.LimitTokensInDaily != 0 {
		t.Errorf("LimitTokensInDaily = %d, want 0 (default)", got.LimitTokensInDaily)
	}
	if got.LimitTokensOutDaily != 0 {
		t.Errorf("LimitTokensOutDaily = %d, want 0 (default)", got.LimitTokensOutDaily)
	}
	if got.TokensInDailyCount != 0 {
		t.Errorf("TokensInDailyCount = %d, want 0 (default)", got.TokensInDailyCount)
	}
	if got.TokensOutDailyCount != 0 {
		t.Errorf("TokensOutDailyCount = %d, want 0 (default)", got.TokensOutDailyCount)
	}
}

func TestListAPIKeysReturnsTokenFields(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k1 := &model.APIKey{
		ID:                  "key-list-1",
		Name:                "list-key-1",
		Key:                 "sk-share-list-1",
		LimitDaily:          200,
		LimitTokensInDaily:  10000,
		LimitTokensOutDaily: 20000,
		AllowedModels:       "",
		Active:              true,
		CreatedAt:           time.Now().UTC().Truncate(time.Second),
	}
	k2 := &model.APIKey{
		ID:                  "key-list-2",
		Name:                "list-key-2",
		Key:                 "sk-share-list-2",
		LimitDaily:          0,
		LimitTokensInDaily:  0,
		LimitTokensOutDaily: 0,
		AllowedModels:       "",
		Active:              true,
		CreatedAt:           time.Now().UTC().Truncate(time.Second),
	}
	if err := s.CreateAPIKey(k1); err != nil {
		t.Fatalf("CreateAPIKey k1: %v", err)
	}
	if err := s.CreateAPIKey(k2); err != nil {
		t.Fatalf("CreateAPIKey k2: %v", err)
	}

	keys, err := s.ListAPIKeys()
	if err != nil {
		t.Fatalf("ListAPIKeys: %v", err)
	}
	if len(keys) < 2 {
		t.Fatalf("ListAPIKeys returned %d keys, want at least 2", len(keys))
	}

	var found1, found2 bool
	for _, k := range keys {
		if k.ID == "key-list-1" {
			found1 = true
			if k.LimitTokensInDaily != 10000 {
				t.Errorf("key-list-1 LimitTokensInDaily = %d, want 10000", k.LimitTokensInDaily)
			}
			if k.LimitTokensOutDaily != 20000 {
				t.Errorf("key-list-1 LimitTokensOutDaily = %d, want 20000", k.LimitTokensOutDaily)
			}
			if k.TokensInDailyCount != 0 {
				t.Errorf("key-list-1 TokensInDailyCount = %d, want 0", k.TokensInDailyCount)
			}
			if k.TokensOutDailyCount != 0 {
				t.Errorf("key-list-1 TokensOutDailyCount = %d, want 0", k.TokensOutDailyCount)
			}
		}
		if k.ID == "key-list-2" {
			found2 = true
			if k.LimitTokensInDaily != 0 {
				t.Errorf("key-list-2 LimitTokensInDaily = %d, want 0", k.LimitTokensInDaily)
			}
			if k.LimitTokensOutDaily != 0 {
				t.Errorf("key-list-2 LimitTokensOutDaily = %d, want 0", k.LimitTokensOutDaily)
			}
		}
	}
	if !found1 {
		t.Error("key-list-1 not found in ListAPIKeys result")
	}
	if !found2 {
		t.Error("key-list-2 not found in ListAPIKeys result")
	}
}

func TestIncrementTokenCount(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-inc-1", "sk-share-tok-inc-1")

	if err := s.IncrementTokenCount(k.ID, 100, 50); err != nil {
		t.Fatalf("IncrementTokenCount() error: %v", err)
	}

	var tokensIn, tokensOut int
	err := s.db.QueryRow(`SELECT tokens_in_daily_count, tokens_out_daily_count FROM api_keys WHERE id = ?`, k.ID).Scan(&tokensIn, &tokensOut)
	if err != nil {
		t.Fatalf("query token counts: %v", err)
	}
	if tokensIn != 100 {
		t.Errorf("tokens_in_daily_count = %d, want %d", tokensIn, 100)
	}
	if tokensOut != 50 {
		t.Errorf("tokens_out_daily_count = %d, want %d", tokensOut, 50)
	}
}

func TestIncrementTokenCountAccumulates(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-acc-1", "sk-share-tok-acc-1")

	if err := s.IncrementTokenCount(k.ID, 100, 50); err != nil {
		t.Fatalf("first IncrementTokenCount() error: %v", err)
	}
	if err := s.IncrementTokenCount(k.ID, 200, 75); err != nil {
		t.Fatalf("second IncrementTokenCount() error: %v", err)
	}

	var tokensIn, tokensOut int
	err := s.db.QueryRow(`SELECT tokens_in_daily_count, tokens_out_daily_count FROM api_keys WHERE id = ?`, k.ID).Scan(&tokensIn, &tokensOut)
	if err != nil {
		t.Fatalf("query token counts: %v", err)
	}
	if tokensIn != 300 {
		t.Errorf("tokens_in_daily_count = %d, want %d", tokensIn, 300)
	}
	if tokensOut != 125 {
		t.Errorf("tokens_out_daily_count = %d, want %d", tokensOut, 125)
	}
}

func TestIncrementTokenCountZeroNoop(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-zero-1", "sk-share-tok-zero-1")

	if err := s.IncrementTokenCount(k.ID, 50, 25); err != nil {
		t.Fatalf("seed IncrementTokenCount() error: %v", err)
	}

	if err := s.IncrementTokenCount(k.ID, 0, 0); err != nil {
		t.Fatalf("IncrementTokenCount(0,0) error: %v", err)
	}

	var tokensIn, tokensOut int
	err := s.db.QueryRow(`SELECT tokens_in_daily_count, tokens_out_daily_count FROM api_keys WHERE id = ?`, k.ID).Scan(&tokensIn, &tokensOut)
	if err != nil {
		t.Fatalf("query token counts: %v", err)
	}
	if tokensIn != 50 {
		t.Errorf("tokens_in_daily_count = %d, want %d (unchanged after no-op)", tokensIn, 50)
	}
	if tokensOut != 25 {
		t.Errorf("tokens_out_daily_count = %d, want %d (unchanged after no-op)", tokensOut, 25)
	}
}

func TestIncrementDailyCount(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-dc-1", "sk-share-tok-dc-1")

	for i := 0; i < 3; i++ {
		if err := s.IncrementDailyCount(k.ID); err != nil {
			t.Fatalf("IncrementDailyCount() call %d error: %v", i+1, err)
		}
	}

	_, dailyCount, err := s.GetDailyCount(k.ID)
	if err != nil {
		t.Fatalf("GetDailyCount: %v", err)
	}
	if dailyCount != 3 {
		t.Errorf("daily_count = %d, want %d", dailyCount, 3)
	}
}

func TestDailyResetClearsAllCounters(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-reset-1", "sk-share-tok-reset-1")

	s.IncrementDailyCountBy(k.ID, 10)
	s.IncrementTokenCount(k.ID, 5000, 3000)

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	if _, err := s.db.Exec(`UPDATE api_keys SET daily_date = ? WHERE id = ?`, yesterday, k.ID); err != nil {
		t.Fatalf("set stale daily_date: %v", err)
	}

	var dailyCount, tokensIn, tokensOut int
	err := s.db.QueryRow(`SELECT daily_count, tokens_in_daily_count, tokens_out_daily_count FROM api_keys WHERE id = ?`, k.ID).
		Scan(&dailyCount, &tokensIn, &tokensOut)
	if err != nil {
		t.Fatalf("pre-check query: %v", err)
	}
	if dailyCount != 10 {
		t.Fatalf("pre: daily_count = %d, want 10", dailyCount)
	}
	if tokensIn != 5000 {
		t.Fatalf("pre: tokens_in_daily_count = %d, want 5000", tokensIn)
	}
	if tokensOut != 3000 {
		t.Fatalf("pre: tokens_out_daily_count = %d, want 3000", tokensOut)
	}

	today := time.Now().UTC().Format("2006-01-02")
	if err := s.ResetDailyCount(k.ID, today); err != nil {
		t.Fatalf("ResetDailyCount: %v", err)
	}

	var storedDate string
	err = s.db.QueryRow(`SELECT daily_count, tokens_in_daily_count, tokens_out_daily_count, daily_date FROM api_keys WHERE id = ?`, k.ID).
		Scan(&dailyCount, &tokensIn, &tokensOut, &storedDate)
	if err != nil {
		t.Fatalf("post-check query: %v", err)
	}
	if dailyCount != 0 {
		t.Errorf("daily_count = %d, want 0 after reset", dailyCount)
	}
	if tokensIn != 0 {
		t.Errorf("tokens_in_daily_count = %d, want 0 after reset", tokensIn)
	}
	if tokensOut != 0 {
		t.Errorf("tokens_out_daily_count = %d, want 0 after reset", tokensOut)
	}
	if len(storedDate) < 10 || storedDate[:10] != today {
		t.Errorf("daily_date = %q, want %q", storedDate, today)
	}
}

func TestIncrementDailyCountOnlyAffectsDaily(t *testing.T) {
	s := MustTestStore()
	defer s.Close()

	k := InsertTestAPIKey(s, "tok-dco-1", "sk-share-tok-dco-1")

	for i := 0; i < 3; i++ {
		if err := s.IncrementDailyCount(k.ID); err != nil {
			t.Fatalf("IncrementDailyCount() call %d error: %v", i+1, err)
		}
	}

	_, dailyCount, err := s.GetDailyCount(k.ID)
	if err != nil {
		t.Fatalf("GetDailyCount: %v", err)
	}
	if dailyCount != 3 {
		t.Errorf("daily_count = %d, want %d", dailyCount, 3)
	}
}
