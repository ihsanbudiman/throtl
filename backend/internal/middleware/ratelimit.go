package middleware

import (
	"fmt"
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

func (rl *RateLimiter) Check(keyID string, limitCount, limitWindowHrs, dailyLimit int) (bool, int, string) {
	now := time.Now()

	// --- Rolling window check ---
	if limitCount > 0 && limitWindowHrs > 0 {
		windowStart, windowCount, err := rl.store.GetRateLimitState(keyID)
		if err != nil {
			return true, 0, ""
		}

		windowDuration := time.Duration(limitWindowHrs) * time.Hour
		if windowStart.IsZero() || now.Sub(windowStart) > windowDuration {
			rl.store.ResetRateLimitWindow(keyID, now)
			windowCount = 0
			windowStart = now
		}

		if windowCount >= limitCount {
			retryAfter := int(windowDuration.Seconds()) - int(now.Sub(windowStart).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			return false, retryAfter, fmt.Sprintf("%d-hour rate limit exceeded", limitWindowHrs)
		}
	}

	// --- Daily limit check (resets at 00:00) ---
	if dailyLimit > 0 {
		today := now.Format("2006-01-02")
		dailyDate, dailyCount, err := rl.store.GetDailyCount(keyID)
		if err != nil {
			return true, 0, ""
		}

		if dailyDate != today {
			rl.store.ResetDailyCount(keyID, today)
			dailyCount = 0
		}

		if dailyCount >= dailyLimit {
			tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			retryAfter := int(tomorrow.Sub(now).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			return false, retryAfter, "Daily rate limit exceeded"
		}
	}

	rl.store.IncrementWindowCount(keyID)
	return true, 0, ""
}

type KeyRateLimitStatus struct {
	Count      int        `json:"count"`
	Limit      int        `json:"limit"`
	WindowHrs  int        `json:"window_hrs"`
	ResetAt    *time.Time `json:"reset_at,omitempty"`
	DailyCount int        `json:"daily_count"`
	DailyLimit int        `json:"daily_limit"`
	DailyReset *time.Time `json:"daily_reset,omitempty"`
}

func (rl *RateLimiter) GetStatus(keyID string) KeyRateLimitStatus {
	key, _ := rl.store.GetAPIKeyByID(keyID)
	if key == nil {
		return KeyRateLimitStatus{}
	}

	status := KeyRateLimitStatus{
		DailyLimit: key.LimitDaily,
		Limit:      key.LimitWindow,
	}

	windowHrs := key.LimitWindowHrs
	if windowHrs == 0 && key.LimitWindow > 0 {
		windowHrs = 5
	}

	windowStart, count, err := rl.store.GetRateLimitState(keyID)
	if err == nil && !windowStart.IsZero() {
		resetTime := windowStart.Add(time.Duration(windowHrs) * time.Hour)
		if time.Now().After(resetTime) {
			count = 0
		}

		status.Count = count
		status.WindowHrs = windowHrs
		status.ResetAt = &resetTime
	} else if key.LimitWindow > 0 {
		status.WindowHrs = windowHrs
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	dailyDate, dailyCount, err := rl.store.GetDailyCount(keyID)
	if err == nil {
		if dailyDate != today {
			dailyCount = 0
		}
		status.DailyCount = dailyCount

		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
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
			limitCount, _ := c.Get("throtl_limit_window").(int)
			limitWindowHrs, _ := c.Get("throtl_limit_window_hrs").(int)
			dailyLimit, _ := c.Get("throtl_limit_daily").(int)

			allowed, retryAfter, reason := rl.Check(keyID, limitCount, limitWindowHrs, dailyLimit)
			if !allowed {
				c.Response().Header().Set("Retry-After", time.Now().Add(time.Duration(retryAfter)*time.Second).Format(time.RFC1123))
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"message":              reason,
						"type":                  "rate_limit_error",
						"retry_after_seconds":  retryAfter,
					},
				})
			}
			return next(c)
		}
	}
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
			c.Set("throtl_limit_window", key.LimitWindow)
			c.Set("throtl_limit_daily", key.LimitDaily)
			c.Set("throtl_limit_window_hrs", key.LimitWindowHrs)
			c.Set("throtl_allowed_models", key.AllowedModels)
			c.Set("throtl_key_obj", key)

			return next(c)
		}
	}
}
