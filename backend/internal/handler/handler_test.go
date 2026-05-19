package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ihsanbudiman/throtl/internal/middleware"
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

func newTestHandler(t *testing.T) (*store.Store, *Handler) {
	t.Helper()
	s, err := store.NewTestStore()
	if err != nil {
		t.Fatalf("NewTestStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	rl := middleware.NewRateLimiter(s)
	h := New(s, "test-secret", rl)
	return s, h
}

func setupAuthContext(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-admin")
	return c, rec
}

type keyResponse struct {
	model.APIKey
	RateLimit rateLimitStatus `json:"rate_limit"`
}

type rateLimitStatus struct {
	DailyCount          int `json:"daily_count"`
	DailyLimit          int `json:"daily_limit"`
	DailyTokensInCount  int `json:"daily_tokens_in_count"`
	DailyTokensInLimit  int `json:"daily_tokens_in_limit"`
	DailyTokensOutCount int `json:"daily_tokens_out_count"`
	DailyTokensOutLimit int `json:"daily_tokens_out_limit"`
}

func TestListAPIKeysIncludesTokenFields(t *testing.T) {
	s, h := newTestHandler(t)

	store.InsertTestProvider(s, "prov-test")

	k := &model.APIKey{
		ID:                  "key-token-1",
		Name:                "token-test-key",
		Key:                 "sk-share-token-test",
		LimitDaily:          500,
		LimitTokensInDaily:  10000,
		LimitTokensOutDaily: 20000,
		AllowedModels:       "",
		Active:              true,
	}
	if err := s.CreateAPIKey(k); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	e := echo.New()
	c, rec := setupAuthContext(e, http.MethodGet, "/api/keys", "")

	if err := h.ListAPIKeys(c); err != nil {
		t.Fatalf("ListAPIKeys returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var result []keyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least 1 key in response")
	}

	rl := result[0].RateLimit
	if rl.DailyTokensInLimit != 10000 {
		t.Errorf("daily_tokens_in_limit = %d, want 10000", rl.DailyTokensInLimit)
	}
	if rl.DailyTokensOutLimit != 20000 {
		t.Errorf("daily_tokens_out_limit = %d, want 20000", rl.DailyTokensOutLimit)
	}
	if rl.DailyTokensInCount != 0 {
		t.Errorf("daily_tokens_in_count = %d, want 0", rl.DailyTokensInCount)
	}
	if rl.DailyTokensOutCount != 0 {
		t.Errorf("daily_tokens_out_count = %d, want 0", rl.DailyTokensOutCount)
	}
	if rl.DailyLimit != 500 {
		t.Errorf("daily_limit = %d, want 500", rl.DailyLimit)
	}
}

func TestListAPIKeysTokenFieldsHaveZeroValuesWhenUnset(t *testing.T) {
	s, h := newTestHandler(t)

	store.InsertTestProvider(s, "prov-test-2")

	k := &model.APIKey{
		ID:            "key-no-token-1",
		Name:          "no-token-key",
		Key:           "sk-share-no-token",
		LimitDaily:    100,
		AllowedModels: "",
		Active:        true,
	}
	if err := s.CreateAPIKey(k); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	e := echo.New()
	c, rec := setupAuthContext(e, http.MethodGet, "/api/keys", "")

	if err := h.ListAPIKeys(c); err != nil {
		t.Fatalf("ListAPIKeys returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var result []keyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least 1 key in response")
	}

	rl := result[0].RateLimit
	if rl.DailyTokensInLimit != 0 {
		t.Errorf("daily_tokens_in_limit = %d, want 0 (unset)", rl.DailyTokensInLimit)
	}
	if rl.DailyTokensOutLimit != 0 {
		t.Errorf("daily_tokens_out_limit = %d, want 0 (unset)", rl.DailyTokensOutLimit)
	}
}

func TestCreateAPIKeyReturnsTokenLimits(t *testing.T) {
	_, h := newTestHandler(t)

	body := `{"name":"test-key","limit_daily":500,"limit_tokens_in_daily":10000,"limit_tokens_out_daily":20000,"allowed_models":""}`

	e := echo.New()
	c, rec := setupAuthContext(e, http.MethodPost, "/api/keys", body)

	if err := h.CreateAPIKey(c); err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var result model.APIKey
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result.Name != "test-key" {
		t.Errorf("name = %q, want %q", result.Name, "test-key")
	}
	if result.LimitTokensInDaily != 10000 {
		t.Errorf("limit_tokens_in_daily = %d, want 10000", result.LimitTokensInDaily)
	}
	if result.LimitTokensOutDaily != 20000 {
		t.Errorf("limit_tokens_out_daily = %d, want 20000", result.LimitTokensOutDaily)
	}
	if result.LimitDaily != 500 {
		t.Errorf("limit_daily = %d, want 500", result.LimitDaily)
	}
}

func TestCreateAPIKeyReturnsZeroTokenLimitsWhenUnset(t *testing.T) {
	_, h := newTestHandler(t)

	body := `{"name":"minimal-key","limit_daily":100}`

	e := echo.New()
	c, rec := setupAuthContext(e, http.MethodPost, "/api/keys", body)

	if err := h.CreateAPIKey(c); err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var result model.APIKey
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result.LimitTokensInDaily != 0 {
		t.Errorf("limit_tokens_in_daily = %d, want 0", result.LimitTokensInDaily)
	}
	if result.LimitTokensOutDaily != 0 {
		t.Errorf("limit_tokens_out_daily = %d, want 0", result.LimitTokensOutDaily)
	}
}
