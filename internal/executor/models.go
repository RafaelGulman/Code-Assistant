package executor

import (
	"app/codeAssistant/internal/lm_settings"
	"app/codeAssistant/internal/scheduling"
)

type Iterator struct {
	id         int
	client     lm_settings.LMStudioClient
	sysPrompt  scheduling.SystemPrompt
	basePrompt scheduling.BasePrompt
}

func NewIterator(newClient lm_settings.LMStudioClient,
	newSysPrompt scheduling.SystemPrompt,
	newBasePrompt scheduling.BasePrompt) *Iterator {
	return &Iterator{
		id:         0,
		client:     newClient,
		sysPrompt:  newSysPrompt,
		basePrompt: newBasePrompt,
	}
}
