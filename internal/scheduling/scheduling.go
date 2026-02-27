package scheduling

import (
	"fmt"
	"strings"
	"time"
)

// NewSystemPrompt создает новый системный промпт с настройками по умолчанию
func NewSystemPrompt() *SystemPrompt {
	return &SystemPrompt{
		CoreInstruction: "Ты - интеллектуальный агент, работающий на локальной машине. " +
			"Твоя задача - помогать пользователю с программированием и системными задачами. " +
			"Используй доступные инструменты для выполнения запросов.",

		AvailableTools: []Tool{
			{
				Name:        "write_code",
				Description: "Написать код на указанном языке программирования",
				Parameters:  []string{"language", "filename", "code"},
				Example:     `{"tool": "write_code", "parameters": {"language": "go", "filename": "main.go", "code": "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"}}`,
			},
			{
				Name:        "test_code",
				Description: "Протестировать код и вернуть результаты",
				Parameters:  []string{"language", "code", "test_cases"},
				Example:     `{"tool": "test_code", "parameters": {"language": "go", "code": "...", "test_cases": "..."}}`,
			},
			{
				Name:        "execute_command",
				Description: "Выполнить терминальную команду",
				Parameters:  []string{"command", "working_dir"},
				Example:     `{"tool": "execute_command", "parameters": {"command": "ls -la", "working_dir": "/home/user"}}`,
			},
			{
				Name:        "read_file",
				Description: "Прочитать содержимое файла",
				Parameters:  []string{"filename", "encoding"},
				Example:     `{"tool": "read_file", "parameters": {"filename": "config.json", "encoding": "utf-8"}}`,
			},
			{
				Name:        "write_file",
				Description: "Записать данные в файл",
				Parameters:  []string{"filename", "content", "mode"},
				Example:     `{"tool": "write_file", "parameters": {"filename": "output.txt", "content": "data", "mode": "overwrite"}}`,
			},
		},

		Constraints: []string{
			"Всегда используй только доступные инструменты для выполнения задач",
			"Не пытайся выполнять команды напрямую без использования инструментов",
			"При ошибках предоставляй подробную информацию о проблеме",
			"Сохраняй все результаты в понятном формате",
			"Проверяй синтаксис кода перед выполнением",
		},

		ResponseFormat: "JSON с полями: tool (название инструмента), parameters (параметры вызова), explanation (объяснение действий)",

		ContextInfo: ContextInfo{
			CurrentTime:    time.Now(),
			WorkingDir:     ".",
			AvailableTools: []string{"write_code", "test_code", "execute_command", "read_file", "write_file"},
			OS:             "Windows 11",
		},
	}
}

/*
Description: Функция создает системный промпт на основании принятых данных
Input: 	Общая_инструкция string,  Инструменты_доступные_ИИ []Tool, ограничения_и_запреты []string,

	формат_ответа(json, md, txt и т.д.) string, контекстная_информация ContextInfo

Output: системный_промпт *SystemPrompt
*/
func NewSystemPromptSet(coreInstruct string, tools []Tool, scopes []string, formatFeedBack string, contextInfo ContextInfo) *SystemPrompt {
	return &SystemPrompt{
		CoreInstruction: coreInstruct,

		AvailableTools: tools,

		Constraints: scopes,

		ResponseFormat: formatFeedBack,

		ContextInfo: contextInfo,
	}
}

// NewBasePrompt создает новый промпт с настройками по умолчанию
func NewBasePrompt(userQuery string) *BasePrompt {
	return &BasePrompt{
		UserQuery:           userQuery,
		ConversationHistory: []Message{},
		AgentState: AgentState{
			ActiveTool:     "",
			LastToolResult: "",
			ErrorCount:     0,
			Variables:      make(map[string]string),
		},
		Priority: "normal",
	}
}

/*
Description: Функция создает промпт на основании принятых данных
Input: 	Общая_инструкция string,  история_запросов []Message, состояния_задач AgentState,

	приоритет string[стоп, в процессе и т.д.]

Output: системный_промпт *SystemPrompt
*/
func NewBasePromptSet(userQuery string, talkHistory []Message, agentState AgentState, prior string) *BasePrompt {
	return &BasePrompt{
		UserQuery:           userQuery,
		ConversationHistory: talkHistory,
		AgentState:          agentState,
		Priority:            prior,
	}
}

// AddMessage добавляет сообщение в историю разговора ??
func (bp *BasePrompt) AddMessage(role, content string) {
	bp.ConversationHistory = append(bp.ConversationHistory, Message{
		Role:    role, //??
		Content: content,
		Time:    time.Now(),
	})
}

// SetToolCall устанавливает информацию о вызове инструмента
func (bp *BasePrompt) SetToolCall(toolName string, params map[string]string, result string) {
	lastIdx := len(bp.ConversationHistory) - 1
	if lastIdx >= 0 {
		bp.ConversationHistory[lastIdx].Tool = &ToolCall{
			ToolName:   toolName,
			Parameters: params,
			Result:     result,
		}
	}
}

// UpdateAgentState обновляет состояние агента
func (bp *BasePrompt) UpdateAgentState(tool string, result string, errCount int) {
	bp.AgentState.ActiveTool = tool
	bp.AgentState.LastToolResult = result
	bp.AgentState.ErrorCount = errCount
}

// BuildPromptString формирует строку промпта для отправки в LM Studio
func (sp *SystemPrompt) BuildPromptString() string {
	var sb strings.Builder

	sb.WriteString("=== СИСТЕМНЫЕ ИНСТРУКЦИИ ===\n")
	sb.WriteString(sp.CoreInstruction + "\n\n")

	sb.WriteString("=== ДОСТУПНЫЕ ИНСТРУМЕНТЫ ===\n")
	for _, tool := range sp.AvailableTools {
		sb.WriteString(fmt.Sprintf("Инструмент: %s\n", tool.Name))
		sb.WriteString(fmt.Sprintf("Описание: %s\n", tool.Description))
		sb.WriteString(fmt.Sprintf("Параметры: %v\n", tool.Parameters))
		sb.WriteString(fmt.Sprintf("Пример: %s\n\n", tool.Example))
	}

	sb.WriteString("=== ОГРАНИЧЕНИЯ ===\n")
	for i, constraint := range sp.Constraints {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, constraint))
	}
	sb.WriteString("\n")

	sb.WriteString("=== ФОРМАТ ОТВЕТА ===\n")
	sb.WriteString(sp.ResponseFormat + "\n\n")

	sb.WriteString("=== КОНТЕКСТ ===\n")
	sb.WriteString(fmt.Sprintf("Текущее время: %s\n", sp.ContextInfo.CurrentTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Рабочая директория: %s\n", sp.ContextInfo.WorkingDir))
	sb.WriteString(fmt.Sprintf("ОС: %s\n", sp.ContextInfo.OS))
	sb.WriteString(fmt.Sprintf("Доступные инструменты: %v\n", sp.ContextInfo.AvailableTools))

	return sb.String()
}

// getOSInfo определяет информацию о операционной системе
func getOSInfo() string {
	// Здесь можно добавить более детальную информацию о системе
	return "Windows" // Заменить на реальное определение
}
