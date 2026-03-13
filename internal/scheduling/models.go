package scheduling

import "time"

// SystemPrompt представляет системный промпт для ИИ агента
type SystemPrompt struct {
	// Основная инструкция для агента
	CoreInstruction string `json:"core_instruction"`

	// Доступные инструменты/опции агента
	AvailableTools []Tool `json:"available_tools"`

	// Ограничения и правила поведения
	Constraints []string `json:"constraints"`

	// Формат ответа
	ResponseFormat string `json:"response_format"`

	// Контекстная информация
	ContextInfo ContextInfo `json:"context_info"`
}

type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
	Example     string   `json:"example"`
}

// ContextInfo содержит контекстную информацию для агента
type ContextInfo struct {
	CurrentTime    time.Time `json:"current_time"`
	WorkingDir     string    `json:"working_dir"`
	AvailableTools []string  `json:"available_tools"`
	OS             string    `json:"os"`
}

// BasePrompt представляет базовый промпт для пользовательских запросов
type BasePrompt struct {
	// Пользовательский запрос
	UserQuery string `json:"user_query"`

	// Контекст предыдущих взаимодействий
	ConversationHistory []Message `json:"conversation_history"`

	// Текущее состояние агента
	AgentState AgentState `json:"agent_state"`

	// Приоритет выполнения
	Priority string `json:"priority"`
}

// Message представляет сообщение в истории разговора
type Message struct {
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
	Tool    *ToolCall `json:"tool,omitempty"`
}

// ToolCall представляет вызов инструмента
type ToolCall struct {
	ToolName   string            `json:"tool_name"`
	Parameters map[string]string `json:"parameters"`
	Result     string            `json:"result"`
}

type AgentState struct {
	ActiveTool     string            `json:"active_tool"`
	LastToolResult string            `json:"last_tool_result"`
	ErrorCount     int               `json:"error_count"`
	Variables      map[string]string `json:"variables"`
}

type TaskStatus int

const (
	stop    TaskStatus = 0
	process TaskStatus = 1
	finish  TaskStatus = 2
	error   TaskStatus = -1
)

type Task struct {
	Description  string
	beginTask    time.Time
	endTask      time.Time
	changeStatus time.Time
	Status       TaskStatus
}

func NewTask(description string) *Task {
	return &Task{
		Description:  description,
		beginTask:    time.Now(),
		changeStatus: time.Now(),
		Status:       process,
	}
}
