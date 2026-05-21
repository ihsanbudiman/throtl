package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

type OpenAIAdapter struct {
	store       *store.Store
	upstream    *http.Client
	modelClient *http.Client
}

func NewOpenAIAdapter(s *store.Store) *OpenAIAdapter {
	return &OpenAIAdapter{
		store:       s,
		upstream:    &http.Client{Timeout: 300 * time.Second},
		modelClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *OpenAIAdapter) DoUpstream(ctx context.Context, provider *model.Provider, body []byte, actualModel string, _ bool) (*ProxyResponse, error) {
	baseURL := strings.TrimRight(provider.BaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	baseURL = strings.TrimRight(baseURL, "/")

	upstreamURL := baseURL + "/v1/chat/completions"

	upstreamReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	upstreamReq = upstreamReq.WithContext(ctx)
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

	upstreamResp, err := a.upstream.Do(upstreamReq)
	if err != nil {
		return nil, err
	}

	return &ProxyResponse{Raw: upstreamResp}, nil
}

func (a *OpenAIAdapter) ListModels(c echo.Context, provider *model.Provider, disabledSet map[string]bool, allowedSet map[string]bool) ([]ModelEntry, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + "/models"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := a.modelClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch models from %s: %v", provider.ID, err)
		return nil, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	var listResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, nil
	}

	var entries []ModelEntry
	for _, m := range listResp.Data {
		prefixedID := provider.ID + "/" + m.ID
		if disabledSet[prefixedID] {
			continue
		}
		if allowedSet != nil && !allowedSet[m.ID] {
			continue
		}
		entries = append(entries, ModelEntry{
			ID:      prefixedID,
			Object:  "model",
			Created: m.Created,
			OwnedBy: provider.ID,
			Slug:    prefixedID,
			DisplayName: m.ID,
			Visibility: "list",
			SupportedInAPI: true,
			DefaultReasoningLevel: "medium",
			SupportedReasoningLevels: []map[string]interface{}{
				{"effort": "low", "description": "Fast responses with lighter reasoning"},
				{"effort": "medium", "description": "Balances speed and reasoning depth"},
				{"effort": "high", "description": "Greater reasoning depth for complex problems"},
			},
			Description: "AI language model",
			ShellType:   "shell_command",
			Priority:    1,
			TruncationPolicy: map[string]interface{}{
				"mode":  "bytes",
				"limit": 10000,
			},
			SupportsParallelToolCalls: true,
			ContextWindow:             128000,
			ExperimentalSupportedTools: []interface{}{},
			InputModalities:            []string{"text", "image"},
		})
	}
	return entries, nil
}

func (a *OpenAIAdapter) FetchModels(provider *model.Provider) ([]UpstreamModel, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + "/models"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := a.modelClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var listResp struct {
		Data []struct {
			ID      string `json:"id"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	models := make([]UpstreamModel, 0, len(listResp.Data))
	for _, m := range listResp.Data {
		models = append(models, UpstreamModel{
			ID:      m.ID,
			Created: m.Created,
			OwnedBy: m.OwnedBy,
		})
	}
	return models, nil
}
