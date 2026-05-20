package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/ihsanbudiman/throtl/internal/middleware"
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/proxy"
	"github.com/ihsanbudiman/throtl/internal/store"
)

type Handler struct {
	store     *store.Store
	jwtSecret []byte
	rl        RateLimitStatusProvider
}

type RateLimitStatusProvider interface {
	GetStatus(keyID string) middleware.KeyRateLimitStatus
}

func New(s *store.Store, jwtSecret string, rl RateLimitStatusProvider) *Handler {
	return &Handler{
		store:     s,
		jwtSecret: []byte(jwtSecret),
		rl:        rl,
	}
}

// --- Auth Handlers ---

func (h *Handler) CheckSetup(c echo.Context) error {
	hasAdmin, err := h.store.HasAdmin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"setup_required": !hasAdmin})
}

func (h *Handler) Setup(c echo.Context) error {
	// Only allow if no admin exists yet
	hasAdmin, err := h.store.HasAdmin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if hasAdmin {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin account already exists"})
	}

	var req model.SetupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email, password, and name are required"})
	}
	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password must be at least 8 characters"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user := &model.User{
		ID:        uuid.New().String()[:8],
		Email:     req.Email,
		Password:  string(hash),
		Name:      req.Name,
		CreatedAt: time.Now(),
	}
	if err := h.store.CreateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusCreated, model.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
	}

	user, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, model.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *Handler) GetMe(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid session"})
	}
	user, err := h.store.GetUserByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

// --- Provider Handlers ---

func (h *Handler) ListProviders(c echo.Context) error {
	providers, err := h.store.ListProviders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	// Mask API keys in response
	for i := range providers {
		providers[i].APIKey = maskKey(providers[i].APIKey)
	}
	return c.JSON(http.StatusOK, providers)
}

func (h *Handler) CreateProvider(c echo.Context) error {
	var req model.CreateProviderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if req.ID == "" || req.Name == "" || req.BaseURL == "" || req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id, name, base_url, and api_key are required"})
	}

	if req.Type != "openai" && req.Type != "anthropic" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "type must be 'openai' or 'anthropic'"})
	}

	if !isValidID(req.ID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id must be lowercase alphanumeric with hyphens (e.g. wafer, openai-us)"})
	}

	p := &model.Provider{
		ID:        req.ID,
		Name:      req.Name,
		Type:      req.Type,
		BaseURL:   strings.TrimRight(req.BaseURL, "/"),
		APIKey:    req.APIKey,
		CreatedAt: time.Now(),
	}
	if err := h.store.CreateProvider(p); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	go func() {
		gw := proxy.NewGateway(h.store)
		models, err := gw.FetchModelsForProvider(p)
		if err != nil {
			log.Printf("Failed to fetch models from %s on creation: %v", p.ID, err)
			return
		}
		for _, m := range models {
			override := &model.ModelOverride{
				ID:         p.ID + "/" + m.ID,
				ProviderID: p.ID,
				ModelName:  m.ID,
				Active:     true,
				CreatedAt:  time.Now(),
			}
			if err := h.store.UpsertModelOverride(override); err != nil {
				log.Printf("Failed to upsert model %s/%s: %v", p.ID, m.ID, err)
			}
		}
	}()

	p.APIKey = maskKey(p.APIKey)
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) DeleteProvider(c echo.Context) error {
	id := c.Param("id")
	if err := h.store.DeleteModelOverridesByProvider(id); err != nil {
		log.Printf("Failed to delete model overrides for provider %s: %v", id, err)
	}
	if err := h.store.DeleteProvider(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// --- API Key Handlers ---

func (h *Handler) ListAPIKeys(c echo.Context) error {
	keys, err := h.store.ListAPIKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	type keyWithStatus struct {
		model.APIKey
		RateLimit middleware.KeyRateLimitStatus `json:"rate_limit"`
	}
	result := make([]keyWithStatus, len(keys))
	for i := range keys {
		keys[i].Key = maskKey(keys[i].Key)
		status := h.rl.GetStatus(keys[i].ID)
		result[i] = keyWithStatus{APIKey: keys[i], RateLimit: status}
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateAPIKey(c echo.Context) error {
	var req model.CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	// Generate share key
	shareKey := "sk-share-" + uuid.New().String()

	k := &model.APIKey{
		ID:                  uuid.New().String()[:8],
		Name:                req.Name,
		Key:                 shareKey,
		LimitDaily:          req.LimitDaily,
		LimitTokensInDaily:  req.LimitTokensInDaily,
		LimitTokensOutDaily: req.LimitTokensOutDaily,
		AllowedModels:       req.AllowedModels,
		Active:              true,
		CreatedAt:           time.Now(),
	}

	if err := h.store.CreateAPIKey(k); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return the full key only on creation
	return c.JSON(http.StatusCreated, k)
}

func (h *Handler) ToggleAPIKey(c echo.Context) error {
	id := c.Param("id")
	active := c.QueryParam("active") == "true"
	if err := h.store.ToggleAPIKey(id, active); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"active": active})
}

func (h *Handler) DeleteAPIKey(c echo.Context) error {
	id := c.Param("id")
	if err := h.store.DeleteAPIKey(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ResetAPIKeyLimits(c echo.Context) error {
	id := c.Param("id")
	today := time.Now().UTC().Format("2006-01-02")
	if err := h.store.ResetDailyCount(id, today); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Dashboard ---

func (h *Handler) GetStats(c echo.Context) error {
	stats, err := h.store.GetDashboardStats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetUsageLogs(c echo.Context) error {
	logs, err := h.store.GetRecentLogs(50)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, logs)
}

// maskKey shows only first 8 and last 4 chars
func maskKey(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:8] + "..." + key[len(key)-4:]
}

func isValidID(id string) bool {
	if len(id) == 0 || len(id) > 32 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func (h *Handler) generateToken(userID string) (string, error) {
	return middleware.GenerateToken(userID, h.jwtSecret)
}

type ModelEntry struct {
	ID               string `json:"id"`
	Object           string `json:"object"`
	Created          int64  `json:"created"`
	OwnedBy          string `json:"owned_by"`
	ProviderID       string `json:"provider_id"`
	Active           bool   `json:"active"`
	RequestMultiplier int   `json:"request_multiplier"`
}

func (h *Handler) ListModels(c echo.Context) error {
	providers, err := h.store.ListProviders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list providers"})
	}

	overrides, err := h.store.ListModelOverrides()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list model overrides"})
	}
	overrideMap := make(map[string]model.ModelOverride)
	for _, o := range overrides {
		key := o.ProviderID + "/" + o.ModelName
		overrideMap[key] = o
	}

	gw := proxy.NewGateway(h.store)
	var data []ModelEntry
	for _, provider := range providers {
		models, err := gw.FetchModelsForProvider(&provider)
		if err != nil {
			log.Printf("Failed to fetch models from %s: %v", provider.ID, err)
			continue
		}

		for _, m := range models {
			prefixedID := provider.ID + "/" + m.ID
			entry := ModelEntry{
				ID:               prefixedID,
				Object:           "model",
				Created:          m.Created,
				OwnedBy:          provider.ID,
				ProviderID:       provider.ID,
				Active:           true,
				RequestMultiplier: 1,
			}
			if o, exists := overrideMap[prefixedID]; exists {
				entry.Active = o.Active
				entry.RequestMultiplier = o.RequestMultiplier
			}
			data = append(data, entry)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

func (h *Handler) ToggleModel(c echo.Context) error {
	modelID := c.Param("id")

	parts := strings.SplitN(modelID, "/", 2)
	if len(parts) != 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model ID must be in format provider-id/model-name"})
	}
	providerID, modelName := parts[0], parts[1]

	provider, err := h.store.GetProvider(providerID)
	if err != nil || provider == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Provider not found"})
	}

	var req model.UpdateModelOverrideRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Fall back to query param for backward compatibility
	if req.Active == nil {
		active := c.QueryParam("active") == "true"
		req.Active = &active
	}

	multiplier := 1
	if req.RequestMultiplier != nil && *req.RequestMultiplier > 0 {
		multiplier = *req.RequestMultiplier
	}

	override := &model.ModelOverride{
		ID:                modelID,
		ProviderID:        providerID,
		ModelName:         modelName,
		Active:            *req.Active,
		RequestMultiplier:  multiplier,
		CreatedAt:         time.Now(),
	}
	if err := h.store.UpsertModelOverride(override); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, override)
}
