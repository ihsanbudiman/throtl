package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (a *OpenAIAdapter) ProxyChat(c echo.Context, provider *model.Provider, body []byte, actualModel string, reqModel string, keyID string) error {
	start := time.Now()

	newBody := strings.Replace(string(body), `"model":"`+reqModel+`"`, `"model":"`+actualModel+`"`, 1)
	newBody = strings.Replace(newBody, `"model": "`+reqModel+`"`, `"model": "`+actualModel+`"`, 1)

	if strings.Contains(newBody, `"stream":true`) || strings.Contains(newBody, `"stream": true`) {
		if !strings.Contains(newBody, `"stream_options"`) {
			newBody = strings.Replace(newBody, `"stream":true`, `"stream":true,"stream_options":{"include_usage":true}`, 1)
			newBody = strings.Replace(newBody, `"stream": true`, `"stream": true,"stream_options":{"include_usage":true}`, 1)
		}
	}
	body = []byte(newBody)

	upstreamPath := c.Request().URL.Path
	upstreamPath = strings.TrimPrefix(upstreamPath, "/v1")
	upstreamURL := strings.TrimRight(provider.BaseURL, "/") + upstreamPath
	if c.Request().URL.RawQuery != "" {
		upstreamURL += "?" + c.Request().URL.RawQuery
	}

	upstreamReq, err := http.NewRequest(c.Request().Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to create upstream request"},
		})
	}

	upstreamReq.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	for _, h := range []string{"OpenAI-Organization", "OpenAI-Beta"} {
		if v := c.Request().Header.Get(h); v != "" {
			upstreamReq.Header.Set(h, v)
		}
	}

	upstreamResp, err := a.upstream.Do(upstreamReq)
	if err != nil {
		log.Printf("Upstream request failed: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error": map[string]string{"message": "Upstream provider error"},
		})
	}
	defer upstreamResp.Body.Close()

	for k, vs := range upstreamResp.Header {
		for _, v := range vs {
			c.Response().Header().Add(k, v)
		}
	}

	isStream := upstreamResp.Header.Get("Content-Type") == "text/event-stream" ||
		strings.Contains(upstreamResp.Header.Get("Content-Type"), "text/event-stream")

	if isStream {
		c.Response().WriteHeader(upstreamResp.StatusCode)
		c.Response().Flush()

		flusher, canFlush := c.Response().Writer.(http.Flusher)
		var streamBuf bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, readErr := upstreamResp.Body.Read(buf)
			if n > 0 {
				c.Response().Writer.Write(buf[:n])
				if canFlush {
					flusher.Flush()
				}
				streamBuf.Write(buf[:n])
			}
			if readErr != nil {
				break
			}
		}

		tokensIn, tokensOut := extractStreamTokens(streamBuf.String())
		go func() {
			latency := int(time.Since(start).Milliseconds())
			usageLog := &model.UsageLog{
				ID:        uuid.New().String(),
				APIKeyID:  keyID,
				Provider:  provider.Name,
				Model:     reqModel,
				Status:    upstreamResp.StatusCode,
				TokensIn:  tokensIn,
				TokensOut: tokensOut,
				LatencyMs: latency,
				CreatedAt: time.Now(),
			}
			if err := a.store.CreateUsageLog(usageLog); err != nil {
				log.Printf("Failed to log usage: %v", err)
			}
			_ = a.store.UpdateLastUsed(keyID)
		}()
		return nil
	}

	respBody, err := io.ReadAll(upstreamResp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to read upstream response"},
		})
	}

	latency := int(time.Since(start).Milliseconds())
	tokensIn, tokensOut := extractTokens(respBody)
	usageLog := &model.UsageLog{
		ID:        uuid.New().String(),
		APIKeyID:  keyID,
		Provider:  provider.Name,
		Model:     reqModel,
		Status:    upstreamResp.StatusCode,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		LatencyMs: latency,
		CreatedAt: time.Now(),
	}
	if err := a.store.CreateUsageLog(usageLog); err != nil {
		log.Printf("Failed to log usage: %v", err)
	}

	go func() {
		_ = a.store.UpdateLastUsed(keyID)
	}()

	return c.Blob(upstreamResp.StatusCode, upstreamResp.Header.Get("Content-Type"), respBody)
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
