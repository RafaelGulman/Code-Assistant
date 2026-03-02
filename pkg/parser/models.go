package parser

// Самый внешний ответ
type OuterResponse struct {
	Tool        string `json:"tool"`
	Parameters  any    `json:"parameters"`
	Explanation string `json:"explanation"`
	RawResponse string `json:"raw_response"` // ← это строка с JSON внутри!
	Error       string `json:"error,omitempty"`
}

// Ответ LM Studio API (внутри RawResponse)
type LMStudioResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
