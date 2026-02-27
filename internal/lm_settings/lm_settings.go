package lm_settings

import (
	"app/codeAssistant/internal/scheduling"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (c *LMStudioClient) SendPrompt(systemPrompt *scheduling.SystemPrompt, basePrompt *scheduling.BasePrompt) (*AgentResponse, error) {
	// Формируем полный промпт
	fullPrompt := c.buildFullPrompt(systemPrompt, basePrompt)

	// Создаем запрос к LM Studio API
	request := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt.BuildPromptString(),
			},
			{
				"role":    "user",
				"content": fullPrompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	// Отправляем запрос
	jsonResponse, err := c.sendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	// Парсим ответ
	response := &AgentResponse{}
	if err := json.Unmarshal(jsonResponse, response); err != nil {
		// Если не удалось распарсить как структуру, сохраняем сырой ответ
		response.RawResponse = string(jsonResponse)
		response.Error = fmt.Sprintf("ошибка парсинга ответа: %v", err)
		return response, nil
	}

	response.RawResponse = string(jsonResponse)
	return response, nil
}

func (c *LMStudioClient) buildFullPrompt(systemPrompt *scheduling.SystemPrompt, basePrompt *scheduling.BasePrompt) string {
	var promptBuilder strings.Builder

	// Добавляем пользовательский запрос
	promptBuilder.WriteString("ПОЛЬЗОВАТЕЛЬСКИЙ ЗАПРОС:\n")
	promptBuilder.WriteString(basePrompt.UserQuery + "\n\n")

	// Добавляем историю разговора, если есть
	if len(basePrompt.ConversationHistory) > 0 {
		promptBuilder.WriteString("ИСТОРИЯ РАЗГОВОРА:\n")
		for _, msg := range basePrompt.ConversationHistory {
			promptBuilder.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				msg.Time.Format(time.RFC3339), msg.Role, msg.Content))

			if msg.Tool != nil {
				promptBuilder.WriteString(fmt.Sprintf("  Инструмент: %s\n", msg.Tool.ToolName))
				promptBuilder.WriteString(fmt.Sprintf("  Результат: %s\n", msg.Tool.Result))
			}
		}
		promptBuilder.WriteString("\n")
	}

	// Добавляем состояние агента
	promptBuilder.WriteString("ТЕКУЩЕЕ СОСТОЯНИЕ АГЕНТА:\n")
	promptBuilder.WriteString(fmt.Sprintf("Активный инструмент: %s\n", basePrompt.AgentState.ActiveTool))
	promptBuilder.WriteString(fmt.Sprintf("Последний результат: %s\n", basePrompt.AgentState.LastToolResult))
	promptBuilder.WriteString(fmt.Sprintf("Количество ошибок: %d\n", basePrompt.AgentState.ErrorCount))
	promptBuilder.WriteString(fmt.Sprintf("Приоритет: %s\n", basePrompt.Priority))

	return promptBuilder.String()
}

func (c *LMStudioClient) sendRequest(request map[string]interface{}) ([]byte, error) {
	// Сериализуем запрос в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("сервер вернул статус %d: %s", resp.StatusCode, string(body))
	}

	// Читаем тело ответа
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return response, nil
}

func (c *LMStudioClient) ExecuteTool(response *AgentResponse) (string, error) {
	switch response.Tool {
	case "write_code":
		return c.executeWriteCode(response.Parameters)
	case "test_code":
		return c.executeTestCode(response.Parameters)
	case "execute_command":
		return c.executeCommand(response.Parameters)
	case "read_file":
		return c.executeReadFile(response.Parameters)
	case "write_file":
		return c.executeWriteFile(response.Parameters)
	default:
		return "", fmt.Errorf("неизвестный инструмент: %s", response.Tool)
	}
}

// Пример реализации инструментов (заглушки)
func (c *LMStudioClient) executeWriteCode(params map[string]string) (string, error) {
	// Реализация записи кода в файл
	return "Код успешно записан", nil
}

func (c *LMStudioClient) executeTestCode(params map[string]string) (string, error) {
	// Реализация тестирования кода
	return "Тесты успешно пройдены", nil
}

func (c *LMStudioClient) executeCommand(params map[string]string) (string, error) {
	// Реализация выполнения команды
	return "Команда выполнена успешно", nil
}

func (c *LMStudioClient) executeReadFile(params map[string]string) (string, error) {
	// Реализация чтения файла
	return "Содержимое файла", nil
}

func (c *LMStudioClient) executeWriteFile(params map[string]string) (string, error) {
	// Реализация записи в файл
	return "Данные успешно записаны", nil
}
