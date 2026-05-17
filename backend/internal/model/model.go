package model

import "time"

// User represents an admin account
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`           // bcrypt hash, never sent to client
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// SetupRequest creates the first (and only) admin account
type SetupRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned on login/setup
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Provider represents an upstream AI provider (e.g. OpenAI)
type Provider struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"api_key"` // real upstream key, encrypted at rest later
	CreatedAt time.Time `json:"created_at"`
}

// APIKey is a generated share key for consumers
type APIKey struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Key              string    `json:"key"`
	LimitWindow      int       `json:"limit_window"`
	LimitDaily       int       `json:"limit_daily"`
	LimitWindowHrs   int       `json:"limit_window_hrs"`
	AllowedModels    string    `json:"allowed_models"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
}

// UsageLog tracks a single proxied request
type UsageLog struct {
	ID        string    `json:"id"`
	APIKeyID  string    `json:"api_key_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Status    int       `json:"status"` // HTTP status from upstream
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	LatencyMs int       `json:"latency_ms"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAPIKeyRequest is the input for generating a new key
type CreateAPIKeyRequest struct {
	Name           string `json:"name" validate:"required"`
	LimitWindow      int    `json:"limit_window"`
	LimitDaily       int    `json:"limit_daily"`
	LimitWindowHrs int    `json:"limit_window_hrs"`
	AllowedModels  string `json:"allowed_models"`
}

// CreateProviderRequest is the input for adding a provider
type CreateProviderRequest struct {
	ID      string `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required"`
	BaseURL string `json:"base_url" validate:"required"`
	APIKey  string `json:"api_key" validate:"required"`
}

// DashboardStats is the overview data
type DashboardStats struct {
	TotalKeys       int              `json:"total_keys"`
	ActiveKeys      int              `json:"active_keys"`
	TotalProviders  int              `json:"total_providers"`
	TotalRequests   int              `json:"total_requests"`
	RequestsToday   int              `json:"requests_today"`
	TotalTokensIn   int64            `json:"total_tokens_in"`
	TotalTokensOut  int64            `json:"total_tokens_out"`
	RecentRequests  []UsageLog       `json:"recent_requests"`
	KeyUsage        []KeyUsageStat   `json:"key_usage"`
	ModelBreakdown  []ModelStat      `json:"model_breakdown"`
}

type KeyUsageStat struct {
	KeyName   string `json:"key_name"`
	KeyPrefix string `json:"key_prefix"`
	Requests  int    `json:"requests"`
	TokensIn  int64  `json:"tokens_in"`
	TokensOut int64  `json:"tokens_out"`
}

type ModelStat struct {
	Model    string `json:"model"`
	Requests int    `json:"requests"`
	TokensIn int64  `json:"tokens_in"`
	TokensOut int64 `json:"tokens_out"`
}

type ModelOverride struct {
	ID         string    `json:"id"`
	ProviderID string    `json:"provider_id"`
	ModelName  string    `json:"model_name"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}
