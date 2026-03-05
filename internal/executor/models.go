package executor

import (
	"app/codeAssistant/internal/lm_settings"
	"app/codeAssistant/internal/scheduling"
	"app/codeAssistant/pkg/parser"
	"encoding/json"
	"log"
)

type Iterator struct {
	id         int
	client     lm_settings.LMStudioClient
	sysPrompt  scheduling.SystemPrompt
	basePrompt scheduling.BasePrompt
}

func NewIterator(newClient lm_settings.LMStudioClient, newSysPrompt scheduling.SystemPrompt, newBasePrompt scheduling.BasePrompt) *Iterator {
	return &Iterator{
		id:         0,
		client:     newClient,
		sysPrompt:  newSysPrompt,
		basePrompt: newBasePrompt,
	}
}

func (iter *Iterator) StartIteration() {
	iter.id++
	resp, err := iter.client.SendPrompt(&iter.sysPrompt, &iter.basePrompt)
	if err != nil {
		log.Fatalf("Error sending prompt: %v", err)
	}
	var result parser.ChatResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling ChatResponse: %v", err)
	}
	contentAfterParse, err := parser.ParseContentResponse(&result)
	iter.basePrompt.AddMessage(contentAfterParse.Parameters.Code)
}
