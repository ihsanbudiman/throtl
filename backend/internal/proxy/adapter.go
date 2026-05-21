package proxy

import (
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

// ModelEntry represents a model in the gateway's unified model list.
type ModelEntry struct {
	ID                        string                   `json:"id"`
	Object                    string                   `json:"object"`
	Created                   int64                    `json:"created"`
	OwnedBy                   string                   `json:"owned_by"`
	Active                    bool                     `json:"active"`
	RequestMultiplier         int                      `json:"request_multiplier"`
	Slug                      string                   `json:"slug"`
	DisplayName               string                   `json:"display_name"`
	Visibility                string                   `json:"visibility"`
	SupportedInAPI            bool                     `json:"supported_in_api"`
	DefaultReasoningLevel     string                   `json:"default_reasoning_level"`
	SupportedReasoningLevels  []map[string]interface{} `json:"supported_reasoning_levels"`
	Description               string                   `json:"description"`
	ShellType                 string                   `json:"shell_type"`
	Priority                  int                      `json:"priority"`
	BaseInstructions          string                   `json:"base_instructions"`
	SupportsReasoningSummaries bool                    `json:"supports_reasoning_summaries"`
	SupportVerbosity          bool                     `json:"support_verbosity"`
	DefaultVerbosity          interface{}              `json:"default_verbosity"`
	ApplyPatchToolType        interface{}              `json:"apply_patch_tool_type"`
	TruncationPolicy          map[string]interface{}   `json:"truncation_policy"`
	SupportsParallelToolCalls bool                     `json:"supports_parallel_tool_calls"`
	SupportsImageDetailOriginal bool                   `json:"supports_image_detail_original"`
	ContextWindow             int                      `json:"context_window"`
	ExperimentalSupportedTools []interface{}            `json:"experimental_supported_tools"`
	InputModalities            []string                 `json:"input_modalities"`
	SupportsSearchTool        bool                     `json:"supports_search_tool"`
	Upgrade                   interface{}              `json:"upgrade"`
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
