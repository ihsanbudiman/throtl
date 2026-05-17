package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func New(dbURL string) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			base_url TEXT NOT NULL,
			api_key TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key TEXT NOT NULL UNIQUE,
			limit_window INT DEFAULT 0,
			limit_daily INT DEFAULT 0,
			limit_window_hrs INT DEFAULT 0,
			window_start TIMESTAMP,
			window_count INT DEFAULT 0,
			daily_count INT DEFAULT 0,
			daily_date DATE,
			allowed_models TEXT DEFAULT '',
			active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_used_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS usage_logs (
			id TEXT PRIMARY KEY,
			api_key_id TEXT NOT NULL REFERENCES api_keys(id),
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			status INT DEFAULT 0,
			tokens_in INT DEFAULT 0,
			tokens_out INT DEFAULT 0,
			latency_ms INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_key ON usage_logs(api_key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_time ON usage_logs(created_at)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS model_overrides (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL REFERENCES providers(id),
			model_name TEXT NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider_id, model_name)
		)`,
		// Drop provider_id from api_keys — keys are no longer bound to a single provider.
		// PostgreSQL doesn't support DROP COLUMN IF EXISTS in a simple way, so catch the error.
	}
	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	// Idempotent DROP COLUMN for provider_id — ignore if column doesn't exist
	if _, err := s.db.Exec(`ALTER TABLE api_keys DROP COLUMN provider_id`); err != nil {
	}

	// Idempotent ALTER TABLE for new columns
	if _, err := s.db.Exec(`ALTER TABLE api_keys ADD COLUMN daily_count INT DEFAULT 0`); err != nil {
	}
	if _, err := s.db.Exec(`ALTER TABLE api_keys ADD COLUMN daily_date DATE`); err != nil {
	}

	// Rename limit_5hr → limit_window, limit_24hr → limit_daily
	if _, err := s.db.Exec(`ALTER TABLE api_keys RENAME COLUMN limit_5hr TO limit_window`); err != nil {
	}
	if _, err := s.db.Exec(`ALTER TABLE api_keys RENAME COLUMN limit_24hr TO limit_daily`); err != nil {
	}

	return nil
}

// --- Providers ---

func (s *Store) CreateProvider(p *model.Provider) error {
	_, err := s.db.Exec(`INSERT INTO providers (id, name, base_url, api_key, created_at) VALUES ($1, $2, $3, $4, $5)`,
		p.ID, p.Name, p.BaseURL, p.APIKey, p.CreatedAt)
	return err
}

func (s *Store) ListProviders() ([]model.Provider, error) {
	rows, err := s.db.Query(`SELECT id, name, base_url, api_key, created_at FROM providers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.Provider, 0)
	for rows.Next() {
		var p model.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (s *Store) GetProvider(id string) (*model.Provider, error) {
	var p model.Provider
	err := s.db.QueryRow(`SELECT id, name, base_url, api_key, created_at FROM providers WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIKey, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *Store) DeleteProvider(id string) error {
	_, err := s.db.Exec(`DELETE FROM providers WHERE id = $1`, id)
	return err
}

// --- API Keys ---

func (s *Store) CreateAPIKey(k *model.APIKey) error {
	_, err := s.db.Exec(`INSERT INTO api_keys (id, name, key, limit_window, limit_daily, limit_window_hrs, allowed_models, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		k.ID, k.Name, k.Key, k.LimitWindow, k.LimitDaily, k.LimitWindowHrs, k.AllowedModels, k.Active, k.CreatedAt)
	return err
}

func (s *Store) ListAPIKeys() ([]model.APIKey, error) {
	rows, err := s.db.Query(`SELECT id, name, key, limit_window, limit_daily, limit_window_hrs, allowed_models, active, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.APIKey, 0)
	for rows.Next() {
		var k model.APIKey
		var lastUsed sql.NullTime
		if err := rows.Scan(&k.ID, &k.Name, &k.Key, &k.LimitWindow, &k.LimitDaily, &k.LimitWindowHrs, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		result = append(result, k)
	}
	return result, nil
}

func (s *Store) GetAPIKeyByShareKey(key string) (*model.APIKey, error) {
	var k model.APIKey
	var lastUsed sql.NullTime
	err := s.db.QueryRow(`SELECT id, name, key, limit_window, limit_daily, limit_window_hrs, allowed_models, active, created_at, last_used_at
		FROM api_keys WHERE key = $1`, key).
		Scan(&k.ID, &k.Name, &k.Key, &k.LimitWindow, &k.LimitDaily, &k.LimitWindowHrs, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	return &k, err
}

func (s *Store) GetAPIKeyByID(id string) (*model.APIKey, error) {
	var k model.APIKey
	var lastUsed sql.NullTime
	err := s.db.QueryRow(`SELECT id, name, key, limit_window, limit_daily, limit_window_hrs, allowed_models, active, created_at, last_used_at
		FROM api_keys WHERE id = $1`, id).
		Scan(&k.ID, &k.Name, &k.Key, &k.LimitWindow, &k.LimitDaily, &k.LimitWindowHrs, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	return &k, err
}

func (s *Store) ToggleAPIKey(id string, active bool) error {
	_, err := s.db.Exec(`UPDATE api_keys SET active = $1 WHERE id = $2`, active, id)
	return err
}

func (s *Store) DeleteAPIKey(id string) error {
	_, err := s.db.Exec(`DELETE FROM usage_logs WHERE api_key_id = $1`, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM api_keys WHERE id = $1`, id)
	return err
}

func (s *Store) UpdateLastUsed(id string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

// --- Usage Logs ---

func (s *Store) CreateUsageLog(l *model.UsageLog) error {
	_, err := s.db.Exec(`INSERT INTO usage_logs (id, api_key_id, provider, model, status, tokens_in, tokens_out, latency_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		l.ID, l.APIKeyID, l.Provider, l.Model, l.Status, l.TokensIn, l.TokensOut, l.LatencyMs, l.CreatedAt)
	return err
}

func (s *Store) GetRecentLogs(limit int) ([]model.UsageLog, error) {
	rows, err := s.db.Query(`SELECT id, api_key_id, provider, model, status, tokens_in, tokens_out, latency_ms, created_at
		FROM usage_logs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.UsageLog, 0)
	for rows.Next() {
		var l model.UsageLog
		if err := rows.Scan(&l.ID, &l.APIKeyID, &l.Provider, &l.Model, &l.Status, &l.TokensIn, &l.TokensOut, &l.LatencyMs, &l.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, l)
	}
	return result, nil
}

func (s *Store) GetDashboardStats() (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	// Total keys
	s.db.QueryRow(`SELECT COUNT(*) FROM api_keys`).Scan(&stats.TotalKeys)
	s.db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE active = TRUE`).Scan(&stats.ActiveKeys)
	s.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&stats.TotalProviders)
	s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs`).Scan(&stats.TotalRequests)

	// Today's requests
	s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs WHERE DATE(created_at) = CURRENT_DATE`).Scan(&stats.RequestsToday)

	// Token totals
	s.db.QueryRow(`SELECT COALESCE(SUM(tokens_in), 0) FROM usage_logs`).Scan(&stats.TotalTokensIn)
	s.db.QueryRow(`SELECT COALESCE(SUM(tokens_out), 0) FROM usage_logs`).Scan(&stats.TotalTokensOut)

	// Recent requests
	stats.RecentRequests, _ = s.GetRecentLogs(10)

	// Key usage
	stats.KeyUsage = make([]model.KeyUsageStat, 0)
	rows, err := s.db.Query(`SELECT k.name, SUBSTRING(k.key, 1, 12) as key_prefix, COUNT(l.id) as requests,
		COALESCE(SUM(l.tokens_in), 0), COALESCE(SUM(l.tokens_out), 0)
		FROM api_keys k LEFT JOIN usage_logs l ON k.id = l.api_key_id
		GROUP BY k.id ORDER BY requests DESC LIMIT 5`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u model.KeyUsageStat
			rows.Scan(&u.KeyName, &u.KeyPrefix, &u.Requests, &u.TokensIn, &u.TokensOut)
			stats.KeyUsage = append(stats.KeyUsage, u)
		}
	}

	// Model breakdown
	stats.ModelBreakdown = make([]model.ModelStat, 0)
	rows2, err := s.db.Query(`SELECT model, COUNT(*) as requests,
		COALESCE(SUM(tokens_in), 0), COALESCE(SUM(tokens_out), 0)
		FROM usage_logs GROUP BY model ORDER BY requests DESC LIMIT 5`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var m model.ModelStat
			rows2.Scan(&m.Model, &m.Requests, &m.TokensIn, &m.TokensOut)
			stats.ModelBreakdown = append(stats.ModelBreakdown, m)
		}
	}

	return stats, nil
}

func (s *Store) GetRequestCountSince(apiKeyID string, since time.Time) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs WHERE api_key_id = $1 AND created_at >= $2`,
		apiKeyID, since).Scan(&count)
	return count, err
}

// --- Rate Limit State (persisted in DB) ---

func (s *Store) GetRateLimitState(keyID string) (windowStart time.Time, count int, err error) {
	var ws sql.NullTime
	err = s.db.QueryRow(`SELECT window_start, window_count FROM api_keys WHERE id = $1`, keyID).Scan(&ws, &count)
	if err == sql.ErrNoRows {
		return time.Time{}, 0, nil
	}
	if ws.Valid {
		windowStart = ws.Time
	}
	return
}

func (s *Store) ResetRateLimitWindow(keyID string, start time.Time) error {
	_, err := s.db.Exec(`UPDATE api_keys SET window_start = $1, window_count = 0 WHERE id = $2`, start, keyID)
	return err
}

func (s *Store) IncrementWindowCount(keyID string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET window_count = window_count + 1, daily_count = daily_count + 1 WHERE id = $1`, keyID)
	return err
}

func (s *Store) GetDailyCount(keyID string) (date string, count int, err error) {
	var ns sql.NullString
	err = s.db.QueryRow(`SELECT daily_date, daily_count FROM api_keys WHERE id = $1`, keyID).Scan(&ns, &count)
	if err == sql.ErrNoRows {
		return "", 0, nil
	}
	if ns.Valid {
		date = ns.String[:10]
	}
	return
}

func (s *Store) ResetDailyCount(keyID string, date string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET daily_date = $1::date, daily_count = 0 WHERE id = $2`, date, keyID)
	return err
}

// --- Users ---

func (s *Store) HasAdmin() (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count > 0, err
}

func (s *Store) CreateUser(u *model.User) error {
	_, err := s.db.Exec(`INSERT INTO users (id, email, password, name, created_at) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.Password, u.Name, u.CreatedAt)
	return err
}

func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := s.db.QueryRow(`SELECT id, email, password, name, created_at FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (s *Store) GetUserByID(id string) (*model.User, error) {
	var u model.User
	err := s.db.QueryRow(`SELECT id, email, password, name, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (s *Store) UpsertModelOverride(m *model.ModelOverride) error {
	_, err := s.db.Exec(`INSERT INTO model_overrides (id, provider_id, model_name, active, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider_id, model_name) DO UPDATE SET active = $4`,
		m.ID, m.ProviderID, m.ModelName, m.Active, m.CreatedAt)
	return err
}

func (s *Store) ListModelOverrides() ([]model.ModelOverride, error) {
	rows, err := s.db.Query(`SELECT id, provider_id, model_name, active, created_at FROM model_overrides`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.ModelOverride, 0)
	for rows.Next() {
		var m model.ModelOverride
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.Active, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (s *Store) GetModelOverride(providerID, modelName string) (*model.ModelOverride, error) {
	var m model.ModelOverride
	err := s.db.QueryRow(`SELECT id, provider_id, model_name, active, created_at FROM model_overrides
		WHERE provider_id = $1 AND model_name = $2`, providerID, modelName).
		Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.Active, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (s *Store) DeleteModelOverridesByProvider(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM model_overrides WHERE provider_id = $1`, providerID)
	return err
}