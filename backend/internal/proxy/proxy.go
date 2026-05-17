package proxy

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/ihsanbudiman/throtl/internal/store"
	"github.com/labstack/echo/v4"
)

type Gateway struct {
	store *store.Store
}

func NewGateway(s *store.Store) *Gateway {
	return &Gateway{store: s}
}

func (g *Gateway) ProxyHandler(c echo.Context) error {
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
		providers, err := g.store.ListProviders()
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

	provider, err := g.store.GetProvider(providerID)
	if err != nil || provider == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"message": "Provider '" + providerID + "' not found",
				"type":    "invalid_request_error",
			},
		})
	}

	newBody := strings.Replace(string(body), `"model":"`+reqModel+`"`, `"model":"`+actualModel+`"`, 1)
	newBody = strings.Replace(newBody, `"model": "`+reqModel+`"`, `"model": "`+actualModel+`"`, 1)
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

	override, _ := g.store.GetModelOverride(providerID, actualModel)
	if override != nil && !override.Active {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error": map[string]string{
				"message": "Model " + actualModel + " is currently disabled",
				"type":    "model_disabled",
			},
		})
	}

	// Normalize: if the request is in Anthropic format, convert to OpenAI format
	// so both adapters always receive OpenAI-format input.
	// AnthropicAdapter will then convert OpenAI→Anthropic for the upstream.
	if isAnthropicRequest(c) {
		body = convertAnthropicRequestToOpenAI(body)
		c.Set("throtl_response_format", "anthropic")
	}

	adapter := NewAdapter(provider.Type, g.store)
	return adapter.ProxyChat(c, provider, body, actualModel, reqModel, keyID)
}

func (g *Gateway) ListModels(c echo.Context) error {
	allowedModels, _ := c.Get("throtl_allowed_models").(string)

	providers, err := g.store.ListProviders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to list providers"},
		})
	}

	overrides, _ := g.store.ListModelOverrides()
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

	var data []ModelEntry

	for _, provider := range providers {
		adapter := NewAdapter(provider.Type, g.store)
		entries, err := adapter.ListModels(c, &provider, disabledSet, allowedSet)
		if err != nil {
			log.Printf("Failed to list models from %s: %v", provider.ID, err)
			continue
		}
		data = append(data, entries...)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   data,
	})
}

func (g *Gateway) FetchModelsForProvider(provider *model.Provider) ([]UpstreamModel, error) {
	adapter := NewAdapter(provider.Type, g.store)
	return adapter.FetchModels(provider)
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

func extractStreamTokens(streamData string) (int, int) {
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
