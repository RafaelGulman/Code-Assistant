package parser

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     any     `json:"logprobs"` // или конкретный тип, если не null
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	ToolCalls []any  `json:"tool_calls"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ContentResponse struct {
	Tool       string `json:"tool"`
	Parameters struct {
		Language string `json:"language"`
		Filename string `json:"filename"`
		Code     string `json:"code"`
	} `json:"parameters"`
	Explanation string `json:"explanation"`
}
