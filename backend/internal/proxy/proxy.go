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
	"github.com/labstack/echo/v4"
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
)

type OpenAIProxy struct {
	store       *store.Store
	upstream    *http.Client
	modelClient *http.Client
}

func NewOpenAIProxy(s *store.Store) *OpenAIProxy {
	return &OpenAIProxy{
		store:       s,
		upstream:    &http.Client{Timeout: 300 * time.Second},
		modelClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OpenAIProxy) ProxyHandler(c echo.Context) error {
	start := time.Now()
	keyID, ok := c.Get("throtl_key_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{"message": "Missing API key context", "type": "authentication_error"},
		})
	}
	allowedModels, _ := c.Get("throtl_allowed_models").(string)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{"message": "Failed to read request body"},
		})
	}

	reqModel := extractModel(body)
	if reqModel == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"message": "Model field is required in request body",
				"type":    "invalid_request_error",
			},
		})
	}

	var providerID, actualModel string
	if idx := strings.Index(reqModel, "/"); idx != -1 {
		providerID = reqModel[:idx]
		actualModel = reqModel[idx+1:]
	} else {
		providers, err := p.store.ListProviders()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": map[string]string{"message": "Failed to list providers"},
			})
		}
		availableProviders := make([]string, 0, len(providers))
		for _, pr := range providers {
			availableProviders = append(availableProviders, pr.ID+" ("+pr.Name+")")
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Model name must include provider prefix. Use format: provider-id/model-name. Available providers: " + strings.Join(availableProviders, ", "),
				"type":    "invalid_request_error",
			},
		})
	}

	provider, err := p.store.GetProvider(providerID)
	if err != nil || provider == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"message": "Provider '" + providerID + "' not found",
				"type":    "invalid_request_error",
			},
		})
	}

	// Rewrite model in request body to strip the provider prefix
	// Handle both "model":"value" and "model": "value" JSON formats
	newBody := strings.Replace(string(body), `"model":"`+reqModel+`"`, `"model":"`+actualModel+`"`, 1)
	newBody = strings.Replace(newBody, `"model": "`+reqModel+`"`, `"model": "`+actualModel+`"`, 1)

	// If streaming, inject stream_options to get usage in the final chunk
	if strings.Contains(newBody, `"stream":true`) || strings.Contains(newBody, `"stream": true`) {
		if !strings.Contains(newBody, `"stream_options"`) {
			newBody = strings.Replace(newBody, `"stream":true`, `"stream":true,"stream_options":{"include_usage":true}`, 1)
			newBody = strings.Replace(newBody, `"stream": true`, `"stream": true,"stream_options":{"include_usage":true}`, 1)
		}
	}
	body = []byte(newBody)

	if allowedModels != "" && actualModel != "" {
		allowed := false
		for _, m := range strings.Split(allowedModels, ",") {
			if strings.TrimSpace(m) == actualModel {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error": map[string]string{
					"message": "Model " + actualModel + " is not allowed for this key",
					"type":    "model_not_allowed",
				},
			})
		}
	}

	override, _ := p.store.GetModelOverride(providerID, actualModel)
	if override != nil && !override.Active {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error": map[string]string{
				"message": "Model " + actualModel + " is currently disabled",
				"type":    "model_disabled",
			},
		})
	}

	// Build upstream URL — strip /v1 prefix from path since base URL may already include it
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

	upstreamResp, err := p.upstream.Do(upstreamReq)
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
			if err := p.store.CreateUsageLog(usageLog); err != nil {
				log.Printf("Failed to log usage: %v", err)
			}
			_ = p.store.UpdateLastUsed(keyID)
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
	if err := p.store.CreateUsageLog(usageLog); err != nil {
		log.Printf("Failed to log usage: %v", err)
	}

	go func() {
		_ = p.store.UpdateLastUsed(keyID)
	}()

	return c.Blob(upstreamResp.StatusCode, upstreamResp.Header.Get("Content-Type"), respBody)
}

func (p *OpenAIProxy) ListModels(c echo.Context) error {
	allowedModels, _ := c.Get("throtl_allowed_models").(string)

	providers, err := p.store.ListProviders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to list providers"},
		})
	}

	overrides, _ := p.store.ListModelOverrides()
	disabledSet := make(map[string]bool)
	for _, o := range overrides {
		if !o.Active {
			disabledSet[o.ProviderID+"/"+o.ModelName] = true
		}
	}

	var allowedSet map[string]bool
	if allowedModels != "" {
		allowedSet = make(map[string]bool)
		for _, m := range strings.Split(allowedModels, ",") {
			allowedSet[strings.TrimSpace(m)] = true
		}
	}

	type modelEntry struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	var data []modelEntry

	for _, provider := range providers {
		url := strings.TrimRight(provider.BaseURL, "/") + "/models"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)

		resp, err := p.modelClient.Do(req)
		if err != nil {
			log.Printf("Failed to fetch models from %s: %v", provider.ID, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
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
			continue
		}

		for _, m := range listResp.Data {
			prefixedID := provider.ID + "/" + m.ID
			if disabledSet[prefixedID] {
				continue
			}
			if allowedSet != nil && !allowedSet[m.ID] {
				continue
			}
			data = append(data, modelEntry{
				ID:      prefixedID,
				Object:  "model",
				Created: m.Created,
				OwnedBy: provider.ID,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

func extractModel(body []byte) string {
	s := string(body)
	idx := strings.Index(s, `"model"`)
	if idx == -1 {
		return ""
	}
	afterColon := s[idx+7:]
	colonIdx := strings.Index(afterColon, ":")
	if colonIdx == -1 {
		return ""
	}
	val := afterColon[colonIdx+1:]
	val = strings.TrimLeft(val, " \t\n\r")
	if len(val) == 0 || val[0] != '"' {
		return ""
	}
	val = val[1:]
	end := strings.Index(val, `"`)
	if end == -1 {
		return ""
	}
	return val[:end]
}

func extractTokens(body []byte) (int, int) {
	s := string(body)
	usageIdx := strings.Index(s, `"usage"`)
	if usageIdx == -1 {
		return 0, 0
	}
	usageBlock := s[usageIdx:]

	promptTokens := extractIntField(usageBlock, `"prompt_tokens"`)
	completionTokens := extractIntField(usageBlock, `"completion_tokens"`)

	return promptTokens, completionTokens
}

// extractStreamTokens parses SSE stream data to find the final usage block
// OpenAI-compatible providers send usage in the last chunk when stream_options.include_usage is true
func extractStreamTokens(streamData string) (int, int) {
	// Find the last "usage" occurrence in the stream — the final chunk has the totals
	lastUsage := strings.LastIndex(streamData, `"usage"`)
	if lastUsage == -1 {
		return 0, 0
	}
	usageBlock := streamData[lastUsage:]
	promptTokens := extractIntField(usageBlock, `"prompt_tokens"`)
	completionTokens := extractIntField(usageBlock, `"completion_tokens"`)
	return promptTokens, completionTokens
}

func extractIntField(s, field string) int {
	idx := strings.Index(s, field)
	if idx == -1 {
		return 0
	}
	after := s[idx+len(field):]
	colonIdx := strings.Index(after, ":")
	if colonIdx == -1 {
		return 0
	}
	val := after[colonIdx+1:]
	val = strings.TrimLeft(val, " \t\n\r")
	end := 0
	for end < len(val) && val[end] >= '0' && val[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0
	}
	result := 0
	for i := 0; i < end; i++ {
		result = result*10 + int(val[i]-'0')
	}
	return result
}
