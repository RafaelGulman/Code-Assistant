package lm_settings

import (
	"net/http"
	"time"
)

// LMStudioClient представляет клиент для взаимодействия с LM Studio
type LMStudioClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	model      string
}

// AgentResponse представляет ответ от ИИ агента
type AgentResponse struct {
	Tool        string            `json:"tool"`
	Parameters  map[string]string `json:"parameters"`
	Explanation string            `json:"explanation"`
	RawResponse string            `json:"raw_response"`
	Error       string            `json:"error,omitempty"`
}

func NewLMStudioClient(baseURL, apiKey, model string) *LMStudioClient {
	return &LMStudioClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Увеличенный таймаут для сложных задач
		},
	}
}
