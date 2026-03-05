package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ParseContentResponse извлекает и парсит JSON из ответа ИИ в структуру ContentResponse
func ParseContentResponse(chatResp *ChatResponse) (*ContentResponse, error) {
	if chatResp == nil {
		return nil, errors.New("chat response is nil")
	}

	if len(chatResp.Choices) == 0 {
		return nil, errors.New("no choices in chat response")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if content == "" {
		return nil, errors.New("empty content in message")
	}

	// Извлекаем чистый JSON из markdown-блока (```json ... ``` или просто ``` ... ```)
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("failed to extract JSON from content: %s", truncateContent(content, 100))
	}

	// Парсим в целевую структуру
	var resp ContentResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w. Raw JSON: %s", err, truncateContent(jsonStr, 200))
	}

	// Валидация обязательных полей
	if resp.Tool == "" {
		return nil, errors.New("missing required field 'tool'")
	}
	if resp.Parameters.Language == "" {
		return nil, errors.New("missing required field 'parameters.language'")
	}
	if resp.Parameters.Filename == "" {
		return nil, errors.New("missing required field 'parameters.filename'")
	}
	if resp.Parameters.Code == "" {
		return nil, errors.New("missing required field 'parameters.code'")
	}
	if resp.Explanation == "" {
		return nil, errors.New("missing required field 'explanation'")
	}

	return &resp, nil
}

// extractJSON извлекает чистый JSON из текста с поддержкой markdown-блоков
func extractJSON(text string) string {
	// Регулярное выражение для поиска JSON внутри ```json ... ``` или ``` ... ```
	// (?s) — включает режим DOTALL (точка совпадает с новой строкой)
	re := regexp.MustCompile(`(?s)[\s\S]*?({.*})[\s\S]*`)
	matches := re.FindStringSubmatch(text)

	// Ищем первый непустой матч с валидной структурой {}
	for i := 1; i < len(matches); i++ {
		candidate := strings.TrimSpace(matches[i])
		if candidate != "" && strings.HasPrefix(candidate, "{") && strings.HasSuffix(candidate, "}") {
			return candidate
		}
	}

	// Если не нашли в блоках — пытаемся обработать весь текст как JSON
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		return text
	}

	return ""
}

// truncateContent обрезает длинный текст для вывода в ошибке
func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (c *ContentResponse) String() string {
	return fmt.Sprintf(
		"Tool: %s\nLanguage: %s\nFilename: %s\nExplanation: %s\nCode:\n%s",
		c.Tool,
		c.Parameters.Language,
		c.Parameters.Filename,
		c.Explanation,
		c.Parameters.Code,
	)
}
