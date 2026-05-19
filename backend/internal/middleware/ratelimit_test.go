package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

func newTestLimiter(t *testing.T) (*store.Store, *RateLimiter) {
	t.Helper()
	s, err := store.NewTestStore()
	if err != nil {
		t.Fatalf("NewTestStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s, NewRateLimiter(s)
}

func setupContext(e *echo.Echo, keyID string, dailyLimit, limitTokensIn, limitTokensOut, tokensInCount, tokensOutCount int) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("throtl_key_id", keyID)
	c.Set("throtl_limit_daily", dailyLimit)
	c.Set("throtl_limit_tokens_in_daily", limitTokensIn)
	c.Set("throtl_limit_tokens_out_daily", limitTokensOut)
	c.Set("throtl_tokens_in_daily_count", tokensInCount)
	c.Set("throtl_tokens_out_daily_count", tokensOutCount)
	return c, rec
}

func TestTokenInputLimitExceeded(t *testing.T) {
	_, rl := newTestLimiter(t)

	e := echo.New()
	c, rec := setupContext(e, "key-1", 0, 1000, 0, 1000, 0)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := rl.Middleware()
	handler := mw(next)

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var body struct {
		Error struct {
			Message            string `json:"message"`
			Type               string `json:"type"`
			LimitType          string `json:"limit_type"`
			Limit              int    `json:"limit"`
			Usage              int    `json:"usage"`
			RetryAfterSeconds  int    `json:"retry_after_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error.LimitType != "daily_tokens_in" {
		t.Errorf("limit_type = %q, want %q", body.Error.LimitType, "daily_tokens_in")
	}
	if body.Error.Type != "rate_limit_error" {
		t.Errorf("type = %q, want %q", body.Error.Type, "rate_limit_error")
	}
	if body.Error.Limit != 1000 {
		t.Errorf("limit = %d, want 1000", body.Error.Limit)
	}
	if body.Error.Usage != 1000 {
		t.Errorf("usage = %d, want 1000", body.Error.Usage)
	}
	if body.Error.RetryAfterSeconds <= 0 {
		t.Errorf("retry_after_seconds = %d, want > 0", body.Error.RetryAfterSeconds)
	}
}

func TestTokenOutputLimitExceeded(t *testing.T) {
	_, rl := newTestLimiter(t)

	e := echo.New()
	c, rec := setupContext(e, "key-1", 0, 0, 5000, 0, 5000)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := rl.Middleware()
	handler := mw(next)

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var body struct {
		Error struct {
			Message            string `json:"message"`
			Type               string `json:"type"`
			LimitType          string `json:"limit_type"`
			Limit              int    `json:"limit"`
			Usage              int    `json:"usage"`
			RetryAfterSeconds  int    `json:"retry_after_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error.LimitType != "daily_tokens_out" {
		t.Errorf("limit_type = %q, want %q", body.Error.LimitType, "daily_tokens_out")
	}
	if body.Error.Type != "rate_limit_error" {
		t.Errorf("type = %q, want %q", body.Error.Type, "rate_limit_error")
	}
	if body.Error.Limit != 5000 {
		t.Errorf("limit = %d, want 5000", body.Error.Limit)
	}
	if body.Error.Usage != 5000 {
		t.Errorf("usage = %d, want 5000", body.Error.Usage)
	}
	if body.Error.RetryAfterSeconds <= 0 {
		t.Errorf("retry_after_seconds = %d, want > 0", body.Error.RetryAfterSeconds)
	}
}

func TestTokenUnlimited(t *testing.T) {
	_, rl := newTestLimiter(t)

	e := echo.New()
	c, rec := setupContext(e, "key-1", 0, 0, 0, 999999, 888888)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := rl.Middleware()
	handler := mw(next)

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (unlimited should pass)", rec.Code, http.StatusOK)
	}
}

func TestDailyRequestLimitExceeded(t *testing.T) {
	s, rl := newTestLimiter(t)

	k := store.InsertTestAPIKey(s, "key-daily-1", "sk-share-daily-1")

	today := time.Now().UTC().Format("2006-01-02")
	s.ResetDailyCount(k.ID, today)

	for i := 0; i < k.LimitDaily; i++ {
		s.IncrementDailyCount(k.ID)
	}

	e := echo.New()
	c, rec := setupContext(e, k.ID, k.LimitDaily, 0, 0, 0, 0)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := rl.Middleware()
	handler := mw(next)

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var body struct {
		Error struct {
			LimitType          string `json:"limit_type"`
			Type               string `json:"type"`
			Limit              int    `json:"limit"`
			Usage              int    `json:"usage"`
			RetryAfterSeconds  int    `json:"retry_after_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error.LimitType != "daily_requests" {
		t.Errorf("limit_type = %q, want %q", body.Error.LimitType, "daily_requests")
	}
	if body.Error.Type != "rate_limit_error" {
		t.Errorf("type = %q, want %q", body.Error.Type, "rate_limit_error")
	}
	if body.Error.Limit != k.LimitDaily {
		t.Errorf("limit = %d, want %d", body.Error.Limit, k.LimitDaily)
	}
	if body.Error.Usage != k.LimitDaily {
		t.Errorf("usage = %d, want %d (actual daily count)", body.Error.Usage, k.LimitDaily)
	}
	if body.Error.RetryAfterSeconds <= 0 {
		t.Errorf("retry_after_seconds = %d, want > 0", body.Error.RetryAfterSeconds)
	}
}

func Test429ResponseBodyStructure(t *testing.T) {
	_, rl := newTestLimiter(t)

	e := echo.New()
	c, rec := setupContext(e, "key-1", 0, 200, 0, 200, 0)

	next := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := rl.Middleware()
	handler := mw(next)

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	var body struct {
		Error struct {
			Message            string `json:"message"`
			Type               string `json:"type"`
			LimitType          string `json:"limit_type"`
			Limit              int    `json:"limit"`
			Usage              int    `json:"usage"`
			RetryAfterSeconds  int    `json:"retry_after_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	required := map[string]bool{
		"message":              body.Error.Message == "",
		"type":                 body.Error.Type == "",
		"limit_type":           body.Error.LimitType == "",
		"limit":                body.Error.Limit == 0 && body.Error.LimitType != "",
		"usage":                body.Error.Usage == 0 && body.Error.LimitType != "",
		"retry_after_seconds":  body.Error.RetryAfterSeconds == 0,
	}
	for field, missing := range required {
		if missing {
			t.Errorf("missing or zero value for field %q in 429 response", field)
		}
	}

	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Retry-After header missing from 429 response")
	}
}
