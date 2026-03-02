package main

import (
	"app/codeAssistant/internal/lm_settings"
	"app/codeAssistant/internal/scheduling"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	kasper := lm_settings.NewLMStudioClient("http://127.0.0.1:1234", "", "meta-llama-3.1-8b-instruct")
	sysPrompt := scheduling.NewSystemPrompt()

	basePrompt := scheduling.NewBasePromptSet("Напиши функцию для генерации случайного набора n-ое кол-ва чисел для golang", []scheduling.Message{}, scheduling.AgentState{}, "normal")
	resp, err := kasper.SendPrompt(sysPrompt, basePrompt)
	data, _ := json.Marshal(resp)
	os.WriteFile("test.json", data, 0644)
	fmt.Println("Response", resp)
	fmt.Println("err:", err)
}
