package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

// ProxyResponse contains the complete upstream response data.
// Used by DoUpstream so the caller can retry before writing to client.
type ProxyResponse struct {
	Raw *http.Response
}

// RetryConfig configures automatic retry behavior for upstream requests.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

var DefaultRetryConfig = RetryConfig{
	MaxRetries: 3,
	BaseDelay:  time.Second,
}

// IsRetryableError returns true if the status code or error warrants a retry.
func IsRetryableError(statusCode int, err error) bool {
	if err != nil {
		return true
	}
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode >= 500 && statusCode <= 599 && statusCode != 501 && statusCode != 505 {
		return true
	}
	return false
}

func (rc RetryConfig) retryWithBackoff(ctx context.Context, fn func() (*ProxyResponse, error)) (*ProxyResponse, error) {
	var lastResp *ProxyResponse
	var lastErr error

	for attempt := 0; attempt <= rc.MaxRetries; attempt++ {
		resp, err := fn()
		lastResp = resp
		lastErr = err

		if resp.Raw != nil && !IsRetryableError(resp.Raw.StatusCode, err) {
			return resp, err
		}

		if resp.Raw != nil {
			resp.Raw.Body.Close()
		}

		if attempt == rc.MaxRetries {
			break
		}

		delay := rc.BaseDelay * (1 << uint(attempt))
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}

		if resp.Raw != nil {
			if retryAfter := resp.Raw.Header.Get("Retry-After"); retryAfter != "" {
				if n := parseRetryAfter(retryAfter); n > delay {
					delay = n
				}
			}
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return lastResp, ctx.Err()
		}
	}

	return lastResp, lastErr
}

func parseRetryAfter(value string) time.Duration {
	var seconds int
	if _, err := fmt.Sscanf(value, "%d", &seconds); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay
		}
	}
	return 0
}

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
	DoUpstream(ctx context.Context, provider *model.Provider, body []byte, actualModel string, wantsAnthropicResponse bool) (*ProxyResponse, error)
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
