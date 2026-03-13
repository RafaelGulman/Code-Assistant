package executor

import (
	"app/codeAssistant/pkg/parser"
	"encoding/json"
	"log"
)

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
