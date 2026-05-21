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

const anthropicVersion = "2023-06-01"

// estimateInputTokens extracts all text content from an Anthropic-format request
// and estimates the input token count. Anthropic's streaming API reports
// input_tokens:0 in the message_start event, so we estimate from the request body.
// Uses ~4 chars per token (conservative for English/mixed content).
func estimateInputTokens(body []byte) int {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return 0
	}

	var totalChars int

	if system, ok := req["system"].(string); ok {
		totalChars += len(system)
	}

	if messages, ok := req["messages"].([]interface{}); ok {
		for _, msgRaw := range messages {
			msg, ok := msgRaw.(map[string]interface{})
			if !ok {
				continue
			}
			content, _ := msg["content"]
			switch c := content.(type) {
			case string:
				totalChars += len(c)
			case []interface{}:
				// content blocks: [{"type":"text","text":"..."}]
				for _, blockRaw := range c {
					block, ok := blockRaw.(map[string]interface{})
					if !ok {
						continue
					}
					if text, ok := block["text"].(string); ok {
						totalChars += len(text)
					}
				}
			}
		}
	}

	if tools, ok := req["tools"].([]interface{}); ok {
		toolBytes, _ := json.Marshal(tools)
		totalChars += len(toolBytes)
	}

	if totalChars == 0 {
		return 0
	}

	tokens := totalChars / 4
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}

type AnthropicAdapter struct {
	store       *store.Store
	upstream    *http.Client
	modelClient *http.Client
}

func NewAnthropicAdapter(s *store.Store) *AnthropicAdapter {
	return &AnthropicAdapter{
		store:       s,
		upstream:    &http.Client{Timeout: 300 * time.Second},
		modelClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *AnthropicAdapter) ProxyChat(c echo.Context, provider *model.Provider, body []byte, actualModel string, reqModel string, keyID string) error {
	start := time.Now()

	wantsAnthropicResponse := c.Get("throtl_response_format") == "anthropic"

	if wantsAnthropicResponse {
		return a.proxyChatPassthrough(c, provider, body, reqModel, keyID, start)
	}
	return a.proxyChatWithTransform(c, provider, body, actualModel, reqModel, keyID, start)
}

// proxyChatPassthrough sends the request to Anthropic and returns the raw
// Anthropic response without any format conversion. Used when the client
// sent an Anthropic-format request and expects Anthropic-format back.
func (a *AnthropicAdapter) proxyChatPassthrough(c echo.Context, provider *model.Provider, body []byte, reqModel string, keyID string, start time.Time) error {
	estimatedInputTokens := estimateInputTokens(body)

	anthropicBody := OpenAIToAnthropicRequest(body)

	upstreamURL := strings.TrimRight(provider.BaseURL, "/") + "/v1/messages"

	upstreamReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(anthropicBody))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to create upstream request"},
		})
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("x-api-key", provider.APIKey)
	upstreamReq.Header.Set("anthropic-version", anthropicVersion)

	upstreamResp, err := a.upstream.Do(upstreamReq)
	if err != nil {
		log.Printf("Upstream request failed: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error": map[string]string{"message": "Upstream provider error"},
		})
	}
	defer upstreamResp.Body.Close()

	isStreamResp := upstreamResp.Header.Get("Content-Type") == "text/event-stream" ||
		strings.Contains(upstreamResp.Header.Get("Content-Type"), "text/event-stream")

	if isStreamResp {
		for k, vs := range upstreamResp.Header {
			for _, v := range vs {
				c.Response().Header().Add(k, v)
			}
		}
		c.Response().WriteHeader(upstreamResp.StatusCode)
		c.Response().Flush()

		flusher, canFlush := c.Response().Writer.(http.Flusher)
		var streamBuf bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, readErr := upstreamResp.Body.Read(buf)
			if n > 0 {
				c.Response().Writer.Write(buf[:n])
				streamBuf.Write(buf[:n])
				if canFlush {
					flusher.Flush()
				}
			}
			if readErr != nil {
				break
			}
		}

		inputTokens, outputTokens := extractAnthropicStreamTokensPassthrough(streamBuf.String())
		if inputTokens == 0 {
			inputTokens = estimatedInputTokens
		}
		log.Printf("Anthropic passthrough stream tokens: in=%d out=%d bufLen=%d", inputTokens, outputTokens, streamBuf.Len())

		go func() {
			if inputTokens > 0 || outputTokens > 0 {
				if err := a.store.IncrementTokenCount(keyID, inputTokens, outputTokens); err != nil {
					log.Printf("Failed to increment token count: %v", err)
				}
			}
			latency := int(time.Since(start).Milliseconds())
			usageLog := &model.UsageLog{
				ID:        uuid.New().String(),
				APIKeyID:  keyID,
				Provider:  provider.Name,
				Model:     reqModel,
				Status:    upstreamResp.StatusCode,
				TokensIn:  inputTokens,
				TokensOut: outputTokens,
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

	inputTokens, outputTokens := ExtractAnthropicTokens(respBody)

	latency := int(time.Since(start).Milliseconds())
	usageLog := &model.UsageLog{
		ID:        uuid.New().String(),
		APIKeyID:  keyID,
		Provider:  provider.Name,
		Model:     reqModel,
		Status:    upstreamResp.StatusCode,
		TokensIn:  inputTokens,
		TokensOut: outputTokens,
		LatencyMs: latency,
		CreatedAt: time.Now(),
	}

	go func() {
		if inputTokens > 0 || outputTokens > 0 {
			if err := a.store.IncrementTokenCount(keyID, inputTokens, outputTokens); err != nil {
				log.Printf("Failed to increment token count: %v", err)
			}
		}
		if err := a.store.CreateUsageLog(usageLog); err != nil {
			log.Printf("Failed to log usage: %v", err)
		}
		_ = a.store.UpdateLastUsed(keyID)
	}()

	for k, vs := range upstreamResp.Header {
		for _, v := range vs {
			c.Response().Header().Add(k, v)
		}
	}
	return c.Blob(upstreamResp.StatusCode, upstreamResp.Header.Get("Content-Type"), respBody)
}

// proxyChatWithTransform sends the request to Anthropic and converts the
// response to OpenAI format. Used when the client sent an OpenAI-format request.
func (a *AnthropicAdapter) proxyChatWithTransform(c echo.Context, provider *model.Provider, body []byte, actualModel string, reqModel string, keyID string, start time.Time) error {
	estimatedInputTokens := estimateInputTokens(body)

	anthropicBody := OpenAIToAnthropicRequest(body)

	upstreamURL := strings.TrimRight(provider.BaseURL, "/") + "/v1/messages"

	upstreamReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(anthropicBody))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{"message": "Failed to create upstream request"},
		})
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("x-api-key", provider.APIKey)
	upstreamReq.Header.Set("anthropic-version", anthropicVersion)

	upstreamResp, err := a.upstream.Do(upstreamReq)
	if err != nil {
		log.Printf("Upstream request failed: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"error": map[string]string{"message": "Upstream provider error"},
		})
	}
	defer upstreamResp.Body.Close()

	isStreamResp := upstreamResp.Header.Get("Content-Type") == "text/event-stream" ||
		strings.Contains(upstreamResp.Header.Get("Content-Type"), "text/event-stream")

	if isStreamResp {
		respID := "chatcmpl-" + uuid.New().String()[:8]

		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().WriteHeader(upstreamResp.StatusCode)
		c.Response().Flush()

		flusher, canFlush := c.Response().Writer.(http.Flusher)

		inputTokens, outputTokens := streamAnthropicToOpenAI(
			upstreamResp.Body, c, respID, actualModel, canFlush, flusher,
		)
		if inputTokens == 0 {
			inputTokens = estimatedInputTokens
		}

		go func() {
			if inputTokens > 0 || outputTokens > 0 {
				if err := a.store.IncrementTokenCount(keyID, inputTokens, outputTokens); err != nil {
					log.Printf("Failed to increment token count: %v", err)
				}
			}
			latency := int(time.Since(start).Milliseconds())
			usageLog := &model.UsageLog{
				ID:        uuid.New().String(),
				APIKeyID:  keyID,
				Provider:  provider.Name,
				Model:     reqModel,
				Status:    upstreamResp.StatusCode,
				TokensIn:  inputTokens,
				TokensOut: outputTokens,
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

	inputTokens, outputTokens := ExtractAnthropicTokens(respBody)

	openaiBody := AnthropicToOpenAIResponse(respBody)
	if openaiBody == nil {
		openaiBody = respBody
	}

	latency := int(time.Since(start).Milliseconds())
	usageLog := &model.UsageLog{
		ID:        uuid.New().String(),
		APIKeyID:  keyID,
		Provider:  provider.Name,
		Model:     reqModel,
		Status:    upstreamResp.StatusCode,
		TokensIn:  inputTokens,
		TokensOut: outputTokens,
		LatencyMs: latency,
		CreatedAt: time.Now(),
	}

	go func() {
		if inputTokens > 0 || outputTokens > 0 {
			if err := a.store.IncrementTokenCount(keyID, inputTokens, outputTokens); err != nil {
				log.Printf("Failed to increment token count: %v", err)
			}
		}
		if err := a.store.CreateUsageLog(usageLog); err != nil {
			log.Printf("Failed to log usage: %v", err)
		}
		_ = a.store.UpdateLastUsed(keyID)
	}()

	return c.Blob(upstreamResp.StatusCode, "application/json", openaiBody)
}

// streamAnthropicToOpenAI reads an Anthropic SSE stream from upstream, transforms
// each SSE event to OpenAI format in real-time, and flushes to the client immediately.
// Returns input and output token counts extracted from the stream.
func streamAnthropicToOpenAI(
	upstream io.Reader,
	c echo.Context,
	id string,
	model string,
	canFlush bool,
	flusher http.Flusher,
) (int, int) {
	var inputTokens, outputTokens int
	roleSent := false
	var sseBuf bytes.Buffer
	readBuf := make([]byte, 4096)

	for {
		n, readErr := upstream.Read(readBuf)
		if n > 0 {
			sseBuf.Write(readBuf[:n])
		}

		// Normalize CRLF to LF — Anthropic SSE may use \r\n line endings
		// which would break \n\n event delimiter detection.
		normalized := strings.ReplaceAll(sseBuf.String(), "\r\n", "\n")
		sseBuf.Reset()
		sseBuf.WriteString(normalized)

		// Process all complete SSE events in the buffer.
		// An SSE event ends with \n\n (blank line).
		for {
			eventData := sseBuf.String()
			endIdx := strings.Index(eventData, "\n\n")
			if endIdx == -1 {
				break
			}

			event := eventData[:endIdx]
			sseBuf.Reset()
			sseBuf.WriteString(eventData[endIdx+2:])

			eventType, dataJSON := parseSSEEvent(event)
			if dataJSON == "" {
				continue
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
				continue
			}

			if eventType == "" {
				eventType, _ = data["type"].(string)
			}

			processAnthropicStreamEvent(eventType, data, id, model, c, canFlush, flusher, &inputTokens, &outputTokens, &roleSent)
		}

		if readErr != nil {
			break
		}
	}

	if sseBuf.Len() > 0 {
		remaining := strings.ReplaceAll(sseBuf.String(), "\r\n", "\n")
		eventType, dataJSON := parseSSEEvent(remaining)
		if dataJSON != "" {
			var data map[string]interface{}
			if json.Unmarshal([]byte(dataJSON), &data) == nil {
				if eventType == "" {
					eventType, _ = data["type"].(string)
				}
				processAnthropicStreamEvent(eventType, data, id, model, c, canFlush, flusher, &inputTokens, &outputTokens, &roleSent)
			}
		}
	}

	return inputTokens, outputTokens
}

func processAnthropicStreamEvent(
	eventType string,
	data map[string]interface{},
	id string,
	model string,
	c echo.Context,
	canFlush bool,
	flusher http.Flusher,
	inputTokens *int,
	outputTokens *int,
	roleSent *bool,
) {
	writeAndFlush := func(s string) {
		c.Response().Writer.Write([]byte(s))
		c.Response().Writer.Write([]byte("\n\n"))
		if canFlush {
			flusher.Flush()
		}
	}

	switch eventType {
	case "message_start":
		if msg, ok := data["message"].(map[string]interface{}); ok {
			if usage, ok := msg["usage"].(map[string]interface{}); ok {
				if it, ok := usage["input_tokens"].(float64); ok {
					*inputTokens = int(it)
				}
			}
		}
		if !*roleSent {
			writeAndFlush(buildOpenAIStreamChunk(id, model, "assistant", "", ""))
			*roleSent = true
		}

	case "content_block_start":
		if block, ok := data["content_block"].(map[string]interface{}); ok {
			if blockType, _ := block["type"].(string); blockType == "tool_use" {
				blockID, _ := block["id"].(string)
				blockName, _ := block["name"].(string)
				toolCallObj := map[string]interface{}{
					"index": data["index"],
					"id":    blockID,
					"type":  "function",
					"function": map[string]interface{}{
						"name":      blockName,
						"arguments": "",
					},
				}
				chunk := map[string]interface{}{
					"id":      id,
					"object":  "chat.completion.chunk",
					"created": time.Now().Unix(),
					"model":   model,
					"choices": []interface{}{
						map[string]interface{}{
							"index":         0,
							"delta":         map[string]interface{}{"tool_calls": []interface{}{toolCallObj}},
							"finish_reason": nil,
						},
					},
				}
				b, _ := json.Marshal(chunk)
				writeAndFlush("data: " + string(b))
			}
		}

	case "content_block_delta":
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			deltaType, _ := delta["type"].(string)
			switch deltaType {
			case "text_delta":
				text, _ := delta["text"].(string)
				writeAndFlush(buildOpenAIStreamChunk(id, model, "", text, ""))
			case "input_json_delta":
				partial, _ := delta["partial_json"].(string)
				chunk := map[string]interface{}{
					"id":      id,
					"object":  "chat.completion.chunk",
					"created": time.Now().Unix(),
					"model":   model,
					"choices": []interface{}{
						map[string]interface{}{
							"index": 0,
							"delta": map[string]interface{}{
								"tool_calls": []interface{}{
									map[string]interface{}{
										"index":   data["index"],
										"function": map[string]interface{}{"arguments": partial},
									},
								},
							},
							"finish_reason": nil,
						},
					},
				}
				b, _ := json.Marshal(chunk)
				writeAndFlush("data: " + string(b))
			}
		}

	case "message_delta":
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			stopReason, _ := delta["stop_reason"].(string)
			finishReason := mapStopReason(stopReason)
			writeAndFlush(buildOpenAIStreamChunk(id, model, "", "", finishReason))
		}
		if usage, ok := data["usage"].(map[string]interface{}); ok {
			if ot, ok := usage["output_tokens"].(float64); ok {
				*outputTokens = int(ot)
			}
			if it, ok := usage["input_tokens"].(float64); ok && *inputTokens == 0 {
				*inputTokens = int(it)
			}
			writeAndFlush(buildOpenAIStreamUsageChunk(id, model, *inputTokens, *outputTokens))
		}

	case "message_stop":
		c.Response().Writer.Write([]byte("data: [DONE]\n\n"))
		if canFlush {
			flusher.Flush()
		}

	case "ping":
	}
}

// extractAnthropicStreamTokensPassthrough parses raw Anthropic SSE stream data
// to extract token counts for usage logging (passthrough mode).
func extractAnthropicStreamTokensPassthrough(streamData string) (int, int) {
	var inputTokens, outputTokens int
	normalized := strings.ReplaceAll(streamData, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if dataStr == "" {
			continue
		}
		var data map[string]interface{}
		if json.Unmarshal([]byte(dataStr), &data) != nil {
			continue
		}
		eventType, _ := data["type"].(string)
		switch eventType {
		case "message_start":
			if msg, ok := data["message"].(map[string]interface{}); ok {
				if usage, ok := msg["usage"].(map[string]interface{}); ok {
					if it, ok := usage["input_tokens"].(float64); ok {
						inputTokens = int(it)
					}
				}
			}
		case "message_delta":
			if usage, ok := data["usage"].(map[string]interface{}); ok {
				if ot, ok := usage["output_tokens"].(float64); ok {
					outputTokens = int(ot)
				}
				if it, ok := usage["input_tokens"].(float64); ok && inputTokens == 0 {
					inputTokens = int(it)
				}
			}
		}
	}
	return inputTokens, outputTokens
}

// parseSSEEvent extracts the event type and data from an SSE event block.
// Input is the text between \n\n delimiters (one complete SSE event).
func parseSSEEvent(event string) (eventType string, data string) {
	for _, line := range strings.Split(event, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	return
}

func (a *AnthropicAdapter) ListModels(c echo.Context, provider *model.Provider, disabledSet map[string]bool, allowedSet map[string]bool) ([]ModelEntry, error) {
	models, err := a.FetchModels(provider)
	if err != nil {
		return nil, err
	}

	var entries []ModelEntry
	for _, m := range models {
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

func (a *AnthropicAdapter) FetchModels(provider *model.Provider) ([]UpstreamModel, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + "/v1/models"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", provider.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)

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
			ID        string `json:"id"`
			CreatedAt int64  `json:"created_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	models := make([]UpstreamModel, 0, len(listResp.Data))
	for _, m := range listResp.Data {
		models = append(models, UpstreamModel{
			ID:      m.ID,
			Created: m.CreatedAt,
			OwnedBy: "anthropic",
		})
	}
	return models, nil
}
