package proxy

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// OpenAIToAnthropicRequest converts an OpenAI-format request body to Anthropic format.
func OpenAIToAnthropicRequest(body []byte) []byte {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return body
	}

	anthropic := make(map[string]interface{})

	if model, ok := req["model"].(string); ok {
		anthropic["model"] = model
	}

	messagesRaw, _ := req["messages"].([]interface{})
	var systemParts []string
	var anthropicMsgs []interface{}

	for _, msgRaw := range messagesRaw {
		msg, ok := msgRaw.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)

		switch role {
		case "system":
			if content, ok := msg["content"].(string); ok {
				systemParts = append(systemParts, content)
			}
		case "user", "assistant":
			anthropicMsgs = append(anthropicMsgs, msg)
		case "tool":
			content, _ := msg["content"].(string)
			toolCallID, _ := msg["tool_call_id"].(string)
			toolResult := map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type":       "tool_result",
						"tool_use_id": toolCallID,
						"content":     content,
					},
				},
			}
			anthropicMsgs = append(anthropicMsgs, toolResult)
		}
	}

	if len(systemParts) > 0 {
		anthropic["system"] = strings.Join(systemParts, "\n")
	}
	anthropic["messages"] = anthropicMsgs

	if maxTokens, ok := req["max_tokens"].(float64); ok {
		anthropic["max_tokens"] = int(maxTokens)
	} else {
		anthropic["max_tokens"] = 4096
	}

	if temperature, ok := req["temperature"].(float64); ok {
		anthropic["temperature"] = temperature
	}
	if topP, ok := req["top_p"].(float64); ok {
		anthropic["top_p"] = topP
	}
	if stop, ok := req["stop"]; ok {
		switch v := stop.(type) {
		case string:
			anthropic["stop_sequences"] = []string{v}
		case []interface{}:
			seqs := make([]string, 0, len(v))
			for _, s := range v {
				if str, ok := s.(string); ok {
					seqs = append(seqs, str)
				}
			}
			anthropic["stop_sequences"] = seqs
		}
	}
	if stream, ok := req["stream"].(bool); ok {
		anthropic["stream"] = stream
	}

	if toolsRaw, ok := req["tools"].([]interface{}); ok {
		var anthropicTools []interface{}
		for _, tRaw := range toolsRaw {
			t, ok := tRaw.(map[string]interface{})
			if !ok {
				continue
			}
			fn, _ := t["function"].(map[string]interface{})
			if fn == nil {
				continue
			}
			name, _ := fn["name"].(string)
			desc, _ := fn["description"].(string)
			params, _ := fn["parameters"].(map[string]interface{})
			anthropicTool := map[string]interface{}{
				"name":        name,
				"description": desc,
			}
			if params != nil {
				anthropicTool["input_schema"] = params
			}
			anthropicTools = append(anthropicTools, anthropicTool)
		}
		if len(anthropicTools) > 0 {
			anthropic["tools"] = anthropicTools
		}
	}

	if toolChoice, ok := req["tool_choice"]; ok {
		switch v := toolChoice.(type) {
		case string:
			switch v {
			case "auto":
				anthropic["tool_choice"] = map[string]interface{}{"type": "auto"}
			case "none":
				anthropic["tool_choice"] = map[string]interface{}{"type": "none"}
			case "required":
				anthropic["tool_choice"] = map[string]interface{}{"type": "any"}
			}
		case map[string]interface{}:
			if fn, ok := v["function"].(map[string]interface{}); ok {
				if name, ok := fn["name"].(string); ok {
					anthropic["tool_choice"] = map[string]interface{}{
						"type": "tool",
						"name": name,
					}
				}
			}
		}
	}

	result, err := json.Marshal(anthropic)
	if err != nil {
		return body
	}
	return result
}

// AnthropicToOpenAIResponse converts an Anthropic-format response body to OpenAI format.
func AnthropicToOpenAIResponse(body []byte) []byte {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body
	}

	openAI := make(map[string]interface{})
	openAI["id"] = resp["id"]
	openAI["object"] = "chat.completion"
	openAI["created"] = time.Now().Unix()

	if model, ok := resp["model"].(string); ok {
		openAI["model"] = model
	}

	contentRaw, _ := resp["content"].([]interface{})
	var textParts []string
	var toolCalls []interface{}
	toolCallIdx := 0

	for _, blockRaw := range contentRaw {
		block, ok := blockRaw.(map[string]interface{})
		if !ok {
			continue
		}
		blockType, _ := block["type"].(string)
		switch blockType {
		case "text":
			if text, ok := block["text"].(string); ok {
				textParts = append(textParts, text)
			}
		case "tool_use":
			id, _ := block["id"].(string)
			name, _ := block["name"].(string)
			input, _ := block["input"].(map[string]interface{})
			argsBytes, _ := json.Marshal(input)
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":   id,
				"type": "function",
				"function": map[string]interface{}{
					"name":      name,
					"arguments": string(argsBytes),
				},
			})
			toolCallIdx++
		}
	}

	message := map[string]interface{}{
		"role":    "assistant",
		"content": strings.Join(textParts, ""),
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}

	stopReason, _ := resp["stop_reason"].(string)
	finishReason := mapStopReason(stopReason)

	openAI["choices"] = []interface{}{
		map[string]interface{}{
			"index":         0,
			"message":       message,
			"finish_reason": finishReason,
		},
	}

	usage, _ := resp["usage"].(map[string]interface{})
	if usage != nil {
		inputTokens, _ := usage["input_tokens"].(float64)
		outputTokens, _ := usage["output_tokens"].(float64)
		openAI["usage"] = map[string]interface{}{
			"prompt_tokens":     int(inputTokens),
			"completion_tokens": int(outputTokens),
			"total_tokens":      int(inputTokens + outputTokens),
		}
	}

	result, err := json.Marshal(openAI)
	if err != nil {
		return body
	}
	return result
}

func mapStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	case "stop_sequence":
		return "stop"
	default:
		if reason != "" {
			return reason
		}
		return "stop"
	}
}

// AnthropicStreamToOpenAI converts Anthropic SSE stream data to OpenAI SSE format.
// Returns the converted stream string, input tokens, and output tokens.
func AnthropicStreamToOpenAI(streamData string, id string, model string) (string, int, int) {
	var openaiChunks []string
	var inputTokens, outputTokens int
	roleSent := false

	lines := strings.Split(streamData, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			if strings.HasPrefix(line, "event:") {
				// event lines are processed with their following data line
			}
			continue
		}

		dataStr := strings.TrimPrefix(line, "data:")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
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
			if !roleSent {
				openaiChunks = append(openaiChunks, buildOpenAIStreamChunk(id, model, "assistant", "", ""))
				roleSent = true
			}

		case "content_block_start":
			// no-op for text, but handle tool_use start
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
					openaiChunks = append(openaiChunks, "data: "+string(b))
				}
			}

		case "content_block_delta":
			if delta, ok := data["delta"].(map[string]interface{}); ok {
				deltaType, _ := delta["type"].(string)
				switch deltaType {
				case "text_delta":
					text, _ := delta["text"].(string)
					openaiChunks = append(openaiChunks, buildOpenAIStreamChunk(id, model, "", text, ""))
				case "input_json_delta":
					partial, _ := delta["partial_json"].(string)
					// Emit as tool_calls argument delta
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
											"index":     data["index"],
											"function":   map[string]interface{}{"arguments": partial},
										},
									},
								},
								"finish_reason": nil,
							},
						},
					}
					b, _ := json.Marshal(chunk)
					openaiChunks = append(openaiChunks, "data: "+string(b))
				}
			}

		case "message_delta":
			if delta, ok := data["delta"].(map[string]interface{}); ok {
				stopReason, _ := delta["stop_reason"].(string)
				finishReason := mapStopReason(stopReason)
				openaiChunks = append(openaiChunks, buildOpenAIStreamChunk(id, model, "", "", finishReason))
			}
			if usage, ok := data["usage"].(map[string]interface{}); ok {
				if ot, ok := usage["output_tokens"].(float64); ok {
					outputTokens = int(ot)
				}
				if it, ok := usage["input_tokens"].(float64); ok && inputTokens == 0 {
					inputTokens = int(it)
				}
				openaiChunks = append(openaiChunks, buildOpenAIStreamUsageChunk(id, model, inputTokens, outputTokens))
			}

		case "message_stop":
			openaiChunks = append(openaiChunks, "data: [DONE]")

		case "ping":
			// skip
		}
	}

	return strings.Join(openaiChunks, "\n\n") + "\n\n", inputTokens, outputTokens
}

// ExtractAnthropicTokens extracts token counts from an Anthropic non-streaming response.
func ExtractAnthropicTokens(body []byte) (int, int) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0
	}
	usage, _ := resp["usage"].(map[string]interface{})
	if usage == nil {
		return 0, 0
	}
	inputTokens, _ := usage["input_tokens"].(float64)
	outputTokens, _ := usage["output_tokens"].(float64)
	return int(inputTokens), int(outputTokens)
}

func buildOpenAIStreamChunk(id, model, role, content, finishReason string) string {
	delta := make(map[string]interface{})
	if role != "" {
		delta["role"] = role
	}
	if content != "" {
		delta["content"] = content
	}

	choice := map[string]interface{}{
		"index": 0,
		"delta": delta,
	}
	if finishReason != "" {
		choice["finish_reason"] = finishReason
	} else {
		choice["finish_reason"] = nil
	}

	chunk := map[string]interface{}{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []interface{}{choice},
	}
	b, _ := json.Marshal(chunk)
	return "data: " + string(b)
}

func buildOpenAIStreamUsageChunk(id, model string, promptTokens, completionTokens int) string {
	chunk := map[string]interface{}{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []interface{}{},
		"usage": map[string]interface{}{
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      promptTokens + completionTokens,
		},
	}
	b, _ := json.Marshal(chunk)
	return "data: " + string(b)
}

// convertAnthropicRequestToOpenAI converts an Anthropic-format request to OpenAI format.
func convertAnthropicRequestToOpenAI(body []byte) []byte {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return body
	}

	openai := make(map[string]interface{})

	if model, ok := req["model"].(string); ok {
		openai["model"] = model
	}

	var messages []interface{}

	if system, ok := req["system"].(string); ok && system != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": system,
		})
	}

	if msgs, ok := req["messages"].([]interface{}); ok {
		messages = append(messages, msgs...)
	}
	openai["messages"] = messages

	if maxTokens, ok := req["max_tokens"].(float64); ok {
		openai["max_tokens"] = int(maxTokens)
	}
	if temperature, ok := req["temperature"].(float64); ok {
		openai["temperature"] = temperature
	}
	if topP, ok := req["top_p"].(float64); ok {
		openai["top_p"] = topP
	}
	if stream, ok := req["stream"].(bool); ok {
		openai["stream"] = stream
	}

	if stopSeqs, ok := req["stop_sequences"].([]interface{}); ok {
		strs := make([]string, 0, len(stopSeqs))
		for _, s := range stopSeqs {
			if str, ok := s.(string); ok {
				strs = append(strs, str)
			}
		}
		if len(strs) == 1 {
			openai["stop"] = strs[0]
		} else if len(strs) > 1 {
			openai["stop"] = strs
		}
	}

	if toolsRaw, ok := req["tools"].([]interface{}); ok {
		var openaiTools []interface{}
		for _, tRaw := range toolsRaw {
			t, ok := tRaw.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := t["name"].(string)
			desc, _ := t["description"].(string)
			inputSchema, _ := t["input_schema"].(map[string]interface{})
			openaiTool := map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        name,
					"description": desc,
				},
			}
			if inputSchema != nil {
				openaiTool["function"].(map[string]interface{})["parameters"] = inputSchema
			}
			openaiTools = append(openaiTools, openaiTool)
		}
		if len(openaiTools) > 0 {
			openai["tools"] = openaiTools
		}
	}

	if toolChoice, ok := req["tool_choice"]; ok {
		switch v := toolChoice.(type) {
		case map[string]interface{}:
			choiceType, _ := v["type"].(string)
			switch choiceType {
			case "auto":
				openai["tool_choice"] = "auto"
			case "none":
				openai["tool_choice"] = "none"
			case "any":
				openai["tool_choice"] = "required"
			case "tool":
				if name, ok := v["name"].(string); ok {
					openai["tool_choice"] = map[string]interface{}{
						"type": "function",
						"function": map[string]interface{}{
							"name": name,
						},
					}
				}
			}
		}
	}

	result, err := json.Marshal(openai)
	if err != nil {
		return body
	}
	return result
}

func isAnthropicRequest(c echo.Context) bool {
	if c.Request().Header.Get("anthropic-version") != "" {
		return true
	}
	if strings.HasSuffix(c.Request().URL.Path, "/messages") {
		return true
	}
	return false
}

func extractStringField(body []byte, field string) string {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	val, _ := m[field].(string)
	return val
}
