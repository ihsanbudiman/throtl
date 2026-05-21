package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ihsanbudiman/throtl/internal/model"
	"github.com/labstack/echo/v4"
)

// ResponsesHandler converts OpenAI Responses API requests to Chat Completions,
// proxies upstream, and converts responses back to Responses API format.
func (g *Gateway) ResponsesHandler(c echo.Context) error {
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

	if override != nil && override.RequestMultiplier > 1 {
		c.Set("throtl_request_multiplier", override.RequestMultiplier)
	}

	chatBody := convertResponsesToChatCompletions(body, actualModel)

	var reqMap map[string]interface{}
	_ = json.Unmarshal(body, &reqMap)
	isStream := false
	if s, ok := reqMap["stream"].(bool); ok {
		isStream = s
	}

	baseURL := strings.TrimRight(provider.BaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	baseURL = strings.TrimRight(baseURL, "/")
	upstreamURL := baseURL + "/v1/chat/completions"

	start := time.Now()

	upstreamReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(chatBody))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to create upstream request"},
		})
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	for _, h := range []string{"OpenAI-Organization", "OpenAI-Beta"} {
		if v := c.Request().Header.Get(h); v != "" {
			upstreamReq.Header.Set(h, v)
		}
	}

	client := &http.Client{Timeout: 300 * time.Second}
	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		log.Printf("Upstream request failed: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error": map[string]string{"message": "Upstream provider error"},
		})
	}
	defer upstreamResp.Body.Close()

	if upstreamResp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(upstreamResp.Body)
		var errResp map[string]interface{}
		if json.Unmarshal(respBody, &errResp) == nil {
			return c.JSON(upstreamResp.StatusCode, errResp)
		}
		return c.JSON(upstreamResp.StatusCode, map[string]interface{}{
			"error": map[string]string{"message": string(respBody)},
		})
	}

	if isStream {
		return g.handleResponsesStream(c, upstreamResp, reqModel, keyID, provider, start)
	}

	respBody, err := io.ReadAll(upstreamResp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to read upstream response"},
		})
	}

	latency := int(time.Since(start).Milliseconds())
	tokensIn, tokensOut := extractTokens(respBody)
	go func() {
		if tokensIn > 0 || tokensOut > 0 {
			if err := g.store.IncrementTokenCount(keyID, tokensIn, tokensOut); err != nil {
				log.Printf("Failed to increment token count: %v", err)
			}
		}
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
		if err := g.store.CreateUsageLog(usageLog); err != nil {
			log.Printf("Failed to log usage: %v", err)
		}
		_ = g.store.UpdateLastUsed(keyID)
	}()

	if mult, ok := c.Get("throtl_request_multiplier").(int); ok && mult > 1 {
		if err := g.store.IncrementDailyCountBy(keyID, mult-1); err != nil {
			log.Printf("Failed to apply request multiplier: %v", err)
		}
	}

	responsesBody := convertChatCompletionsToResponses(respBody, reqModel)
	return c.JSONBlob(http.StatusOK, responsesBody)
}

// convertResponsesToChatCompletions converts a Responses API request body to Chat Completions format.
func convertResponsesToChatCompletions(body []byte, actualModel string) []byte {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return body
	}

	chat := make(map[string]interface{})
	chat["model"] = actualModel

	var messages []interface{}

	if instructions, ok := req["instructions"].(string); ok && instructions != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": instructions,
		})
	}

	switch input := req["input"].(type) {
	case string:
		if input != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": input,
			})
		}
	case []interface{}:
		for _, itemRaw := range input {
			item, ok := itemRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if itemType, ok := item["type"].(string); ok {
				switch itemType {
				case "input_text":
					if text, ok := item["text"].(string); ok {
						messages = append(messages, map[string]interface{}{
							"role":    "user",
							"content": text,
						})
					}
					continue
				case "message":
					convertMessageItem(item, &messages)
					continue
				}
			}
			if role, ok := item["role"].(string); ok {
				convertRoleItem(role, item, &messages)
			}
		}
	}

	var mergedSystem string
	var nonSystemMsgs []interface{}
	for _, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			if role, _ := m["role"].(string); role == "system" {
				if content, ok := m["content"].(string); ok {
					if mergedSystem != "" {
						mergedSystem += "\n\n"
					}
					mergedSystem += content
				}
				continue
			}
		}
		nonSystemMsgs = append(nonSystemMsgs, msg)
	}

	messages = nil
	if mergedSystem != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": mergedSystem,
		})
	}
	messages = append(messages, nonSystemMsgs...)

	chat["messages"] = messages

	if maxOutput, ok := req["max_output_tokens"]; ok {
		chat["max_tokens"] = maxOutput
	}

	for _, field := range []string{"temperature", "top_p", "stream", "tool_choice"} {
		if val, ok := req[field]; ok {
			chat[field] = val
		}
	}

	if toolsRaw, ok := req["tools"].([]interface{}); ok {
		var chatTools []interface{}
		for _, tRaw := range toolsRaw {
			t, ok := tRaw.(map[string]interface{})
			if !ok {
				continue
			}
			toolType, _ := t["type"].(string)
			if toolType != "function" {
				continue
			}
			if _, hasFn := t["function"]; hasFn {
				chatTools = append(chatTools, t)
				continue
			}
			fn := map[string]interface{}{}
			if name, ok := t["name"].(string); ok {
				fn["name"] = name
			}
			if desc, ok := t["description"].(string); ok {
				fn["description"] = desc
			}
			if params, ok := t["parameters"]; ok {
				fn["parameters"] = params
			}
			if strict, ok := t["strict"].(bool); ok {
				fn["strict"] = strict
			}
			chatTools = append(chatTools, map[string]interface{}{
				"type":     "function",
				"function": fn,
			})
		}
		if len(chatTools) > 0 {
			chat["tools"] = chatTools
		}
	}

	if textObj, ok := req["text"].(map[string]interface{}); ok {
		if format, ok := textObj["format"]; ok {
			chat["response_format"] = format
		}
	}

	result, err := json.Marshal(chat)
	if err != nil {
		return body
	}
	return result
}

func convertMessageItem(item map[string]interface{}, messages *[]interface{}) {
	role, _ := item["role"].(string)
	content := item["content"]

	mappedRole := role
	if role == "developer" {
		mappedRole = "system"
	}

	switch c := content.(type) {
	case string:
		*messages = append(*messages, map[string]interface{}{
			"role":    mappedRole,
			"content": c,
		})
	case []interface{}:
		var textParts []string
		for _, partRaw := range c {
			part, ok := partRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := part["text"].(string); ok {
				textParts = append(textParts, t)
			}
		}
		if len(textParts) > 0 {
			*messages = append(*messages, map[string]interface{}{
				"role":    mappedRole,
				"content": strings.Join(textParts, ""),
			})
		}
	default:
		*messages = append(*messages, map[string]interface{}{
			"role":    mappedRole,
			"content": content,
		})
	}
}

func convertRoleItem(role string, item map[string]interface{}, messages *[]interface{}) {
	mappedRole := role
	if role == "developer" {
		mappedRole = "system"
	}

	content := item["content"]
	switch c := content.(type) {
	case string:
		*messages = append(*messages, map[string]interface{}{
			"role":    mappedRole,
			"content": c,
		})
	case []interface{}:
		var textParts []string
		for _, partRaw := range c {
			part, ok := partRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := part["text"].(string); ok {
				textParts = append(textParts, t)
			}
		}
		if len(textParts) > 0 {
			*messages = append(*messages, map[string]interface{}{
				"role":    mappedRole,
				"content": strings.Join(textParts, ""),
			})
		}
	default:
		*messages = append(*messages, map[string]interface{}{
			"role":    mappedRole,
			"content": content,
		})
	}
}

// convertChatCompletionsToResponses converts a Chat Completions response to Responses API format.
func convertChatCompletionsToResponses(body []byte, reqModel string) []byte {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body
	}

	respID := "resp_" + uuid.New().String()
	msgID := "msg_" + uuid.New().String()
	contentID := "c_" + uuid.New().String()

	result := map[string]interface{}{
		"id":         respID,
		"object":     "response",
		"created_at": float64(time.Now().Unix()),
		"status":     "completed",
		"model":      reqModel,
	}

	var output []interface{}
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		choice, ok := choices[0].(map[string]interface{})
		if ok {
			message, _ := choice["message"].(map[string]interface{})
			if message != nil {
				var contentText string
				if ct, ok := message["content"].(string); ok {
					contentText = ct
				}

				outputItem := map[string]interface{}{
					"id":   msgID,
					"type": "message",
					"role": "assistant",
					"content": []interface{}{
						map[string]interface{}{
							"id":   contentID,
							"type": "output_text",
							"text": contentText,
						},
					},
				}

				if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
					for _, tcRaw := range toolCalls {
						tc, ok := tcRaw.(map[string]interface{})
						if !ok {
							continue
						}
						tcID, _ := tc["id"].(string)
						fn, _ := tc["function"].(map[string]interface{})
						fnName, _ := fn["name"].(string)
						fnArgs, _ := fn["arguments"].(string)

						var argsObj interface{}
						if json.Unmarshal([]byte(fnArgs), &argsObj) != nil {
							argsObj = fnArgs
						}

						output = append(output, map[string]interface{}{
							"id":        tcID,
							"type":      "function_call",
							"name":      fnName,
							"call_id":   tcID,
							"arguments": argsObj,
						})
					}
				}

				output = append(output, outputItem)
			}
		}
	}

	result["output"] = output

	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		promptTokens, _ := usage["prompt_tokens"].(float64)
		completionTokens, _ := usage["completion_tokens"].(float64)
		result["usage"] = map[string]interface{}{
			"input_tokens":  int(promptTokens),
			"output_tokens": int(completionTokens),
			"total_tokens":  int(promptTokens + completionTokens),
		}
	}

	b, err := json.Marshal(result)
	if err != nil {
		return body
	}
	return b
}

// handleResponsesStream converts a streaming Chat Completions response to Responses API SSE events.
func (g *Gateway) handleResponsesStream(c echo.Context, upstreamResp *http.Response, reqModel string, keyID string, provider *model.Provider, start time.Time) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	flusher, canFlush := c.Response().Writer.(http.Flusher)

	respID := "resp_" + uuid.New().String()
	msgID := "msg_" + uuid.New().String()
	contentID := "c_" + uuid.New().String()
	created := float64(time.Now().Unix())

	writeSSEEvent(c, "response.created", map[string]interface{}{
		"type": "response.created",
		"response": map[string]interface{}{
			"id":         respID,
			"object":     "response",
			"created_at": created,
			"status":     "in_progress",
			"model":      reqModel,
			"output":     []interface{}{},
		},
	}, flusher, canFlush)

	writeSSEEvent(c, "response.output_item.added", map[string]interface{}{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item": map[string]interface{}{
			"id":      msgID,
			"type":    "message",
			"role":    "assistant",
			"content": []interface{}{},
		},
	}, flusher, canFlush)

	writeSSEEvent(c, "response.content_part.added", map[string]interface{}{
		"type":               "response.content_part.added",
		"part_index":         0,
		"item_id":            msgID,
		"content_part_index": 0,
		"part": map[string]interface{}{
			"id":   contentID,
			"type": "output_text",
			"text": "",
		},
	}, flusher, canFlush)

	var fullText strings.Builder
	var tokensIn, tokensOut int
	var streamBuf bytes.Buffer

	scanner := bufio.NewScanner(upstreamResp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		streamBuf.WriteString(line)
		streamBuf.WriteByte('\n')

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		dataStr := strings.TrimPrefix(line, "data:")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr == "[DONE]" {
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
			continue
		}

		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			choice, ok := choices[0].(map[string]interface{})
			if !ok {
				continue
			}
			delta, _ := choice["delta"].(map[string]interface{})
			if delta != nil {
				if content, ok := delta["content"].(string); ok && content != "" {
					fullText.WriteString(content)
					writeSSEEvent(c, "response.output_text.delta", map[string]interface{}{
						"type":               "response.output_text.delta",
						"delta":              content,
						"item_id":            msgID,
						"output_index":       0,
						"content_part_index": 0,
					}, flusher, canFlush)
				}
			}

			if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
				writeSSEEvent(c, "response.content_part.done", map[string]interface{}{
					"type":               "response.content_part.done",
					"part_index":         0,
					"item_id":            msgID,
					"content_part_index": 0,
					"part": map[string]interface{}{
						"id":   contentID,
						"type": "output_text",
						"text": fullText.String(),
					},
				}, flusher, canFlush)

				writeSSEEvent(c, "response.output_item.done", map[string]interface{}{
					"type":         "response.output_item.done",
					"output_index": 0,
					"item": map[string]interface{}{
						"id":   msgID,
						"type": "message",
						"role": "assistant",
						"content": []interface{}{
							map[string]interface{}{
								"id":   contentID,
								"type": "output_text",
								"text": fullText.String(),
							},
						},
					},
				}, flusher, canFlush)
			}
		}

		if usage, ok := chunk["usage"].(map[string]interface{}); ok {
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				tokensIn = int(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				tokensOut = int(ct)
			}
		}
	}

	if tokensIn == 0 && tokensOut == 0 {
		tokensIn, tokensOut = extractStreamTokens(streamBuf.String())
	}

	completedResponse := map[string]interface{}{
		"id":         respID,
		"object":     "response",
		"created_at": created,
		"status":     "completed",
		"model":      reqModel,
		"output": []interface{}{
			map[string]interface{}{
				"id":   msgID,
				"type": "message",
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"id":   contentID,
						"type": "output_text",
						"text": fullText.String(),
					},
				},
			},
		},
	}
	if tokensIn > 0 || tokensOut > 0 {
		completedResponse["usage"] = map[string]interface{}{
			"input_tokens":  tokensIn,
			"output_tokens": tokensOut,
			"total_tokens":  tokensIn + tokensOut,
		}
	}

	writeSSEEvent(c, "response.completed", map[string]interface{}{
		"type":     "response.completed",
		"response": completedResponse,
	}, flusher, canFlush)

	latency := int(time.Since(start).Milliseconds())
	go func() {
		if tokensIn > 0 || tokensOut > 0 {
			if err := g.store.IncrementTokenCount(keyID, tokensIn, tokensOut); err != nil {
				log.Printf("Failed to increment token count: %v", err)
			}
		}
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
		if err := g.store.CreateUsageLog(usageLog); err != nil {
			log.Printf("Failed to log usage: %v", err)
		}
		_ = g.store.UpdateLastUsed(keyID)
	}()

	if mult, ok := c.Get("throtl_request_multiplier").(int); ok && mult > 1 {
		if err := g.store.IncrementDailyCountBy(keyID, mult-1); err != nil {
			log.Printf("Failed to apply request multiplier: %v", err)
		}
	}

	return nil
}

func writeSSEEvent(c echo.Context, eventType string, data interface{}, flusher http.Flusher, canFlush bool) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(c.Response().Writer, "event: %s\ndata: %s\n\n", eventType, string(dataBytes))
	if canFlush {
		flusher.Flush()
	}
}
