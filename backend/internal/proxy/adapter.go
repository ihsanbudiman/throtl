package proxy

import (
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

// ModelEntry represents a model in the gateway's unified model list.
type ModelEntry struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	OwnedBy           string `json:"owned_by"`
	Active            bool   `json:"active"`
	RequestMultiplier int    `json:"request_multiplier"`
}

// UpstreamModel represents a raw model fetched from an upstream provider.
type UpstreamModel struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ProviderAdapter is the interface that each provider type must implement.
type ProviderAdapter interface {
	ProxyChat(c echo.Context, provider *model.Provider, body []byte, actualModel string, reqModel string, keyID string) error
	ListModels(c echo.Context, provider *model.Provider, disabledSet map[string]bool, allowedSet map[string]bool) ([]ModelEntry, error)
	FetchModels(provider *model.Provider) ([]UpstreamModel, error)
}

// NewAdapter returns the appropriate ProviderAdapter for the given provider type.
func NewAdapter(providerType string, s *store.Store) ProviderAdapter {
	switch providerType {
	case "anthropic":
		return NewAnthropicAdapter(s)
	default:
		return NewOpenAIAdapter(s)
	}
}
