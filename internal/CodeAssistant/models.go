
type Task struct {
	ID          string
	Description string
	Status      string // "pending", "in-progress", "done", "failed"
	Result      string
	Code        string
	TestCode    string
}

type Agent struct {
	name         string
	lmStudioURL  string
	workspaceDir string
	taskFile     string
	logFile      string
	maxRetries   int
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}