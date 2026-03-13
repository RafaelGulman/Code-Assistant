package main

import (
	"app/codeAssistant/internal/executor"
	"app/codeAssistant/internal/lm_settings"
	"app/codeAssistant/internal/scheduling"
)

func main() {
	kasper := lm_settings.NewLMStudioClient("http://127.0.0.1:1234", "", "qwen2.5-coder-7b-instruct")
	sysPrompt := scheduling.NewSystemPrompt()

	basePrompt := scheduling.NewBasePromptSet(
		"Напиши функцию которая будет рисовать елку в .txt файле. Используй язык golang",
		[]scheduling.Message{},
		scheduling.AgentState{},
		"normal",
	)
	iter := executor.NewIterator(*kasper, *sysPrompt, *basePrompt)

	iter.StartIteration()

	// resp, err := kasper.SendPrompt(sysPrompt, basePrompt)
	// if err != nil {
	// 	log.Fatalf("Error sending prompt: %v", err)
	// }

	// var result parser.ChatResponse
	// err = json.Unmarshal(resp, &result)
	// if err != nil {
	// 	log.Fatalf("Error unmarshaling ChatResponse: %v", err)
	// }
	// fmt.Println(result)
	// content, err := parser.ParseContentResponse(&result)
	// fmt.Println(content.String())
}
