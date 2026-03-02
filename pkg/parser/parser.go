package parser

import (
	"app/codeAssistant/internal/lm_settings"
	"encoding/json"
	"fmt"
)

func ParseNestedResponse(outerJSON []byte) (*AgentResponse, error) {
	// ШАГ 1: Парсим самый внешний JSON
	var outer OuterResponse
	if err := json.Unmarshal(outerJSON, &outer); err != nil {
		return nil, fmt.Errorf("outer parse: %w", err)
	}

	// ШАГ 2: Парсим raw_response как JSON (LLM API ответ)
	var lmResp LMStudioResponse
	if err := json.Unmarshal([]byte(outer.RawResponse), &lmResp); err != nil {
		return nil, fmt.Errorf("raw_response parse: %w", err)
	}

	if len(lmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	content := lmResp.Choices[0].Message.Content

	// ШАГ 3: Извлекаем JSON из markdown-блока в content
	jsonStr, err := extractJSONFromMarkdown(content)
	if err != nil {
		return nil, fmt.Errorf("extract from markdown: %w", err)
	}

	// ШАГ 4: Чиним двойное экранирование в поле "code"
	jsonStr = fixDoubleEscapedCode(jsonStr)

	// ШАГ 5: Парсим в целевую структуру
	var agent lm_settings.AgentResponse
	if err := json.Unmarshal([]byte(jsonStr), &agent); err != nil {
		return nil, fmt.Errorf("agent parse: %w\nJSON: %s", err, jsonStr)
	}

	// Сохраняем код отдельно для удобства
	if code, ok := agent.Parameters["code"].(string); ok {
		agent.Code = code
	}

	return &agent, nil
}

// extractJSONFromMarkdown находит JSON внутри ```json ... ```
func extractJSONFromMarkdown(content string) (string, error) {
	// Ищем блок ```json { ... } ```
	re := regexp.MustCompile(`(?s)```(?:json)?\s*(\{.*?\})\s*````)
	matches := re.FindStringSubmatch(content)
	
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Фоллбэк: ищем первый { и последний }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end > start {
		return content[start : end+1], nil
	}

	return "", fmt.Errorf("no JSON block found in content")
}

// fixDoubleEscapedCode чинит \\\" → \" внутри поля "code"
func fixDoubleEscapedCode(jsonStr string) string {
	// Находим значение поля "code" и убираем лишние экранирования
	re := regexp.MustCompile(`("code"\s*:\s*)"(.*?)"(?=\s*[,\}])`)
	
	return re.ReplaceAllStringFunc(jsonStr, func(match string) string {
		// Разделяем ключ и значение
		parts := strings.SplitN(match, ":", 2)
		if len(parts) != 2 {
			return match
		}
		
		key := parts[0] + ":"
		value := strings.TrimSpace(parts[1])
		
		// Убираем кавычки по краям значения
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
			// Чиним экранирование: \\\" → \"
			value = strings.ReplaceAll(value, `\"`, `"`)
			// Возвращаем с одинарными кавычками
			return key + `"` + value + `"`
		}
		return match
	})
}
