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
