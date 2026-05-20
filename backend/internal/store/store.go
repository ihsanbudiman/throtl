package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbURL string) (*Store, error) {
	dsn := dbURL + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
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
			type TEXT NOT NULL DEFAULT 'openai',
			base_url TEXT NOT NULL,
			api_key TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key TEXT NOT NULL UNIQUE,
			limit_window INT DEFAULT 0,
			limit_daily INT DEFAULT 0,
			limit_window_hrs INT DEFAULT 0,
			window_start DATETIME,
			window_count INT DEFAULT 0,
			daily_count INT DEFAULT 0,
			daily_date DATE,
			allowed_models TEXT DEFAULT '',
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_key ON usage_logs(api_key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_time ON usage_logs(created_at)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS model_overrides (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL REFERENCES providers(id),
			model_name TEXT NOT NULL,
			active INTEGER DEFAULT 1,
			request_multiplier INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider_id, model_name)
		)`,
	}
	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	// Backward-compatible: add type column to existing databases
	s.db.Exec(`ALTER TABLE providers ADD COLUMN type TEXT NOT NULL DEFAULT 'openai'`)

	// Backward-compatible: add request_multiplier column
	s.db.Exec(`ALTER TABLE model_overrides ADD COLUMN request_multiplier INTEGER DEFAULT 1`)

	// Backward-compatible: add token limit columns
	s.db.Exec(`ALTER TABLE api_keys ADD COLUMN limit_tokens_in_daily INT DEFAULT 0`)
	s.db.Exec(`ALTER TABLE api_keys ADD COLUMN limit_tokens_out_daily INT DEFAULT 0`)
	s.db.Exec(`ALTER TABLE api_keys ADD COLUMN tokens_in_daily_count INT DEFAULT 0`)
	s.db.Exec(`ALTER TABLE api_keys ADD COLUMN tokens_out_daily_count INT DEFAULT 0`)

	return nil
}

// --- Providers ---

func (s *Store) CreateProvider(p *model.Provider) error {
	_, err := s.db.Exec(`INSERT INTO providers (id, name, type, base_url, api_key, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Type, p.BaseURL, p.APIKey, p.CreatedAt)
	return err
}

func (s *Store) ListProviders() ([]model.Provider, error) {
	rows, err := s.db.Query(`SELECT id, name, type, base_url, api_key, created_at FROM providers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.Provider, 0)
	for rows.Next() {
		var p model.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.APIKey, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) GetProvider(id string) (*model.Provider, error) {
	var p model.Provider
	err := s.db.QueryRow(`SELECT id, name, type, base_url, api_key, created_at FROM providers WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.APIKey, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *Store) DeleteProvider(id string) error {
	if err := s.DeleteModelOverridesByProvider(id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM providers WHERE id = ?`, id)
	return err
}

func (s *Store) UpdateProvider(p *model.Provider) error {
	_, err := s.db.Exec(`UPDATE providers SET name = ?, type = ?, base_url = ?, api_key = ? WHERE id = ?`,
		p.Name, p.Type, p.BaseURL, p.APIKey, p.ID)
	return err
}

// --- API Keys ---

func (s *Store) CreateAPIKey(k *model.APIKey) error {
	_, err := s.db.Exec(`INSERT INTO api_keys (id, name, key, limit_daily, limit_tokens_in_daily, limit_tokens_out_daily, tokens_in_daily_count, tokens_out_daily_count, allowed_models, active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.Name, k.Key, k.LimitDaily, k.LimitTokensInDaily, k.LimitTokensOutDaily, k.TokensInDailyCount, k.TokensOutDailyCount, k.AllowedModels, k.Active, k.CreatedAt)
	return err
}

func (s *Store) ListAPIKeys() ([]model.APIKey, error) {
	rows, err := s.db.Query(`SELECT id, name, key, limit_daily, limit_tokens_in_daily, limit_tokens_out_daily, tokens_in_daily_count, tokens_out_daily_count, allowed_models, active, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.APIKey, 0)
	for rows.Next() {
		var k model.APIKey
		var lastUsed sql.NullTime
		if err := rows.Scan(&k.ID, &k.Name, &k.Key, &k.LimitDaily, &k.LimitTokensInDaily, &k.LimitTokensOutDaily, &k.TokensInDailyCount, &k.TokensOutDailyCount, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		result = append(result, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) GetAPIKeyByShareKey(key string) (*model.APIKey, error) {
	var k model.APIKey
	var lastUsed sql.NullTime
	err := s.db.QueryRow(`SELECT id, name, key, limit_daily, limit_tokens_in_daily, limit_tokens_out_daily, tokens_in_daily_count, tokens_out_daily_count, allowed_models, active, created_at, last_used_at
		FROM api_keys WHERE key = ?`, key).
		Scan(&k.ID, &k.Name, &k.Key, &k.LimitDaily, &k.LimitTokensInDaily, &k.LimitTokensOutDaily, &k.TokensInDailyCount, &k.TokensOutDailyCount, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed)
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
	err := s.db.QueryRow(`SELECT id, name, key, limit_daily, limit_tokens_in_daily, limit_tokens_out_daily, tokens_in_daily_count, tokens_out_daily_count, allowed_models, active, created_at, last_used_at
		FROM api_keys WHERE id = ?`, id).
		Scan(&k.ID, &k.Name, &k.Key, &k.LimitDaily, &k.LimitTokensInDaily, &k.LimitTokensOutDaily, &k.TokensInDailyCount, &k.TokensOutDailyCount, &k.AllowedModels, &k.Active, &k.CreatedAt, &lastUsed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	return &k, err
}

func (s *Store) ToggleAPIKey(id string, active bool) error {
	_, err := s.db.Exec(`UPDATE api_keys SET active = ? WHERE id = ?`, active, id)
	return err
}

func (s *Store) DeleteAPIKey(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM usage_logs WHERE api_key_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM api_keys WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateLastUsed(id string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET last_used_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

// --- Usage Logs ---

func (s *Store) CreateUsageLog(l *model.UsageLog) error {
	_, err := s.db.Exec(`INSERT INTO usage_logs (id, api_key_id, provider, model, status, tokens_in, tokens_out, latency_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.APIKeyID, l.Provider, l.Model, l.Status, l.TokensIn, l.TokensOut, l.LatencyMs, l.CreatedAt)
	return err
}

func (s *Store) GetRecentLogs(limit int) ([]model.UsageLog, error) {
	rows, err := s.db.Query(`SELECT id, api_key_id, provider, model, status, tokens_in, tokens_out, latency_ms, created_at
		FROM usage_logs ORDER BY created_at DESC LIMIT ?`, limit)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) GetDashboardStats() (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	// Total keys
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM api_keys`).Scan(&stats.TotalKeys); err != nil {
		log.Printf("store: failed to query total keys: %v", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE active = 1`).Scan(&stats.ActiveKeys); err != nil {
		log.Printf("store: failed to query active keys: %v", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&stats.TotalProviders); err != nil {
		log.Printf("store: failed to query total providers: %v", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs`).Scan(&stats.TotalRequests); err != nil {
		log.Printf("store: failed to query total requests: %v", err)
	}

	// Today's requests
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs WHERE DATE(created_at) = CURRENT_DATE`).Scan(&stats.RequestsToday); err != nil {
		log.Printf("store: failed to query requests today: %v", err)
	}

	// Token totals
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(tokens_in), 0) FROM usage_logs`).Scan(&stats.TotalTokensIn); err != nil {
		log.Printf("store: failed to query total tokens in: %v", err)
	}
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(tokens_out), 0) FROM usage_logs`).Scan(&stats.TotalTokensOut); err != nil {
		log.Printf("store: failed to query total tokens out: %v", err)
	}

	// Recent requests
	stats.RecentRequests, _ = s.GetRecentLogs(10)

	// Key usage
	stats.KeyUsage = make([]model.KeyUsageStat, 0)
	rows, err := s.db.Query(`SELECT k.name, SUBSTR(k.key, 1, 12) as key_prefix, COUNT(l.id) as requests,
		COALESCE(SUM(l.tokens_in), 0), COALESCE(SUM(l.tokens_out), 0)
		FROM api_keys k LEFT JOIN usage_logs l ON k.id = l.api_key_id
		GROUP BY k.id ORDER BY requests DESC LIMIT 5`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u model.KeyUsageStat
			if err := rows.Scan(&u.KeyName, &u.KeyPrefix, &u.Requests, &u.TokensIn, &u.TokensOut); err != nil {
				continue
			}
			stats.KeyUsage = append(stats.KeyUsage, u)
		}
		if err := rows.Err(); err != nil {
			log.Printf("store: key usage rows error: %v", err)
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
			if err := rows2.Scan(&m.Model, &m.Requests, &m.TokensIn, &m.TokensOut); err != nil {
				continue
			}
			stats.ModelBreakdown = append(stats.ModelBreakdown, m)
		}
		if err := rows2.Err(); err != nil {
			log.Printf("store: model breakdown rows error: %v", err)
		}
	}

	return stats, nil
}

func (s *Store) GetRequestCountSince(apiKeyID string, since time.Time) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM usage_logs WHERE api_key_id = ? AND created_at >= ?`,
		apiKeyID, since).Scan(&count)
	return count, err
}

// --- Rate Limit State (persisted in DB) ---

func (s *Store) IncrementDailyCount(keyID string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET daily_count = daily_count + 1 WHERE id = ?`, keyID)
	return err
}

func (s *Store) IncrementDailyCountBy(keyID string, delta int) error {
	if delta <= 0 {
		return nil
	}
	_, err := s.db.Exec(`UPDATE api_keys SET daily_count = daily_count + ? WHERE id = ?`, delta, keyID)
	return err
}

func (s *Store) IncrementTokenCount(keyID string, tokensIn int, tokensOut int) error {
	if tokensIn == 0 && tokensOut == 0 {
		return nil
	}
	_, err := s.db.ExecContext(context.Background(),
		`UPDATE api_keys SET tokens_in_daily_count = tokens_in_daily_count + ?, tokens_out_daily_count = tokens_out_daily_count + ? WHERE id = ?`,
		tokensIn, tokensOut, keyID)
	return err
}

func (s *Store) UpdateModelOverrideMultiplier(id string, multiplier int) error {
	_, err := s.db.Exec(`UPDATE model_overrides SET request_multiplier = ? WHERE id = ?`, multiplier, id)
	return err
}

func (s *Store) GetDailyCount(keyID string) (date string, count int, err error) {
	var ns sql.NullString
	err = s.db.QueryRow(`SELECT daily_date, daily_count FROM api_keys WHERE id = ?`, keyID).Scan(&ns, &count)
	if err == sql.ErrNoRows {
		return "", 0, nil
	}
	if ns.Valid && len(ns.String) >= 10 {
		date = ns.String[:10]
	}
	return
}

func (s *Store) ResetDailyCount(keyID string, date string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET daily_count = 0, tokens_in_daily_count = 0, tokens_out_daily_count = 0, daily_date = ? WHERE id = ?`, date, keyID)
	return err
}

// --- Users ---

func (s *Store) HasAdmin() (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count > 0, err
}

func (s *Store) CreateUser(u *model.User) error {
	_, err := s.db.Exec(`INSERT INTO users (id, email, password, name, created_at) VALUES (?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.Password, u.Name, u.CreatedAt)
	return err
}

func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := s.db.QueryRow(`SELECT id, email, password, name, created_at FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (s *Store) GetUserByID(id string) (*model.User, error) {
	var u model.User
	err := s.db.QueryRow(`SELECT id, email, password, name, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (s *Store) UpsertModelOverride(m *model.ModelOverride) error {
	_, err := s.db.Exec(`INSERT INTO model_overrides (id, provider_id, model_name, active, request_multiplier, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (provider_id, model_name) DO UPDATE SET active = ?, request_multiplier = ?`,
		m.ID, m.ProviderID, m.ModelName, m.Active, m.RequestMultiplier, m.CreatedAt, m.Active, m.RequestMultiplier)
	return err
}

func (s *Store) ListModelOverrides() ([]model.ModelOverride, error) {
	rows, err := s.db.Query(`SELECT id, provider_id, model_name, active, request_multiplier, created_at FROM model_overrides`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.ModelOverride, 0)
	for rows.Next() {
		var m model.ModelOverride
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.Active, &m.RequestMultiplier, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) GetModelOverride(providerID, modelName string) (*model.ModelOverride, error) {
	var m model.ModelOverride
	err := s.db.QueryRow(`SELECT id, provider_id, model_name, active, request_multiplier, created_at FROM model_overrides
		WHERE provider_id = ? AND model_name = ?`, providerID, modelName).
		Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.Active, &m.RequestMultiplier, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (s *Store) DeleteModelOverridesByProvider(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM model_overrides WHERE provider_id = ?`, providerID)
	return err
}
