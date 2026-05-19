package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ihsanbudiman/throtl/internal/store"
)

type RateLimiter struct {
	store *store.Store
}

func NewRateLimiter(s *store.Store) *RateLimiter {
	return &RateLimiter{store: s}
}

func (rl *RateLimiter) Check(keyID string, dailyLimit int) (bool, int, string) {
	now := time.Now()

	// --- Daily limit check (resets at 00:00 UTC) ---
	if dailyLimit > 0 {
		today := now.UTC().Format("2006-01-02")
		dailyDate, dailyCount, err := rl.store.GetDailyCount(keyID)
		if err != nil {
			return true, 0, ""
		}

		if dailyDate != today {
			rl.store.ResetDailyCount(keyID, today)
			dailyCount = 0
		}

		if dailyCount >= dailyLimit {
			tomorrow := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day()+1, 0, 0, 0, 0, time.UTC)
			retryAfter := int(tomorrow.Sub(now).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			return false, retryAfter, "Daily rate limit exceeded"
		}
	}

	rl.store.IncrementDailyCount(keyID)
	return true, 0, ""
}

type KeyRateLimitStatus struct {
	DailyCount          int        `json:"daily_count"`
	DailyLimit          int        `json:"daily_limit"`
	DailyTokensInCount  int        `json:"daily_tokens_in_count"`
	DailyTokensInLimit  int        `json:"daily_tokens_in_limit"`
	DailyTokensOutCount int        `json:"daily_tokens_out_count"`
	DailyTokensOutLimit int        `json:"daily_tokens_out_limit"`
	DailyReset          *time.Time `json:"daily_reset,omitempty"`
}

func (rl *RateLimiter) GetStatus(keyID string) KeyRateLimitStatus {
	key, _ := rl.store.GetAPIKeyByID(keyID)
	if key == nil {
		return KeyRateLimitStatus{}
	}

	status := KeyRateLimitStatus{
		DailyLimit:          key.LimitDaily,
		DailyTokensInLimit:  key.LimitTokensInDaily,
		DailyTokensOutLimit: key.LimitTokensOutDaily,
	}

	now := time.Now()
	today := now.UTC().Format("2006-01-02")
	dailyDate, dailyCount, err := rl.store.GetDailyCount(keyID)
	if err == nil {
		if dailyDate != today {
			dailyCount = 0
		}
		status.DailyCount = dailyCount
		status.DailyTokensInCount = key.TokensInDailyCount
		status.DailyTokensOutCount = key.TokensOutDailyCount

		tomorrow := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day()+1, 0, 0, 0, 0, time.UTC)
		status.DailyReset = &tomorrow
	}

	return status
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			keyID, ok := c.Get("throtl_key_id").(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": map[string]string{"message": "Missing API key context", "type": "authentication_error"},
				})
			}
			dailyLimit, _ := c.Get("throtl_limit_daily").(int)

			allowed, retryAfter, reason := rl.Check(keyID, dailyLimit)
			if !allowed {
				c.Response().Header().Set("Retry-After", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC1123))
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"message":              reason,
						"type":                  "rate_limit_error",
						"limit_type":            "daily_requests",
						"limit":                 dailyLimit,
						"usage":                 dailyLimit,
						"retry_after_seconds":   retryAfter,
					},
				})
			}

			limitTokensIn, _ := c.Get("throtl_limit_tokens_in_daily").(int)
			tokensInCount, _ := c.Get("throtl_tokens_in_daily_count").(int)
			if limitTokensIn > 0 && tokensInCount >= limitTokensIn {
				retryAfter := secondsUntilMidnightUTC()
				c.Response().Header().Set("Retry-After", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC1123))
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"message":              "Daily token input limit exceeded",
						"type":                  "rate_limit_error",
						"limit_type":            "daily_tokens_in",
						"limit":                 limitTokensIn,
						"usage":                 tokensInCount,
						"retry_after_seconds":   retryAfter,
					},
				})
			}

			limitTokensOut, _ := c.Get("throtl_limit_tokens_out_daily").(int)
			tokensOutCount, _ := c.Get("throtl_tokens_out_daily_count").(int)
			if limitTokensOut > 0 && tokensOutCount >= limitTokensOut {
				retryAfter := secondsUntilMidnightUTC()
				c.Response().Header().Set("Retry-After", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC1123))
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"message":              "Daily token output limit exceeded",
						"type":                  "rate_limit_error",
						"limit_type":            "daily_tokens_out",
						"limit":                 limitTokensOut,
						"usage":                 tokensOutCount,
						"retry_after_seconds":   retryAfter,
					},
				})
			}

			return next(c)
		}
	}
}

func secondsUntilMidnightUTC() int {
	now := time.Now().UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	s := int(tomorrow.Sub(now).Seconds())
	if s < 0 {
		return 0
	}
	return s
}

func KeyAuth(s *store.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := ""

			auth := c.Request().Header.Get("Authorization")
			if auth != "" {
				token = auth
				if len(auth) > 7 && auth[:7] == "Bearer " {
					token = auth[7:]
				}
			}

			if token == "" {
				token = c.Request().Header.Get("x-api-key")
			}

			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": map[string]string{
						"message": "Missing Authorization header or x-api-key",
						"type":    "authentication_error",
					},
				})
			}

			key, err := s.GetAPIKeyByShareKey(token)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": map[string]string{"message": "Internal error"},
				})
			}
			if key == nil || !key.Active {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": map[string]string{
						"message": "Invalid API key",
						"type":    "authentication_error",
					},
				})
			}

			c.Set("throtl_key_id", key.ID)
			c.Set("throtl_limit_daily", key.LimitDaily)
			c.Set("throtl_limit_tokens_in_daily", key.LimitTokensInDaily)
			c.Set("throtl_limit_tokens_out_daily", key.LimitTokensOutDaily)
			c.Set("throtl_tokens_in_daily_count", key.TokensInDailyCount)
			c.Set("throtl_tokens_out_daily_count", key.TokensOutDailyCount)
			c.Set("throtl_allowed_models", key.AllowedModels)
			c.Set("throtl_key_obj", key)

			return next(c)
		}
	}
}
