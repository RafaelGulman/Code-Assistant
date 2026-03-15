package main

import (
	"app/codeAssistant/internal/db"
	"app/codeAssistant/internal/executor"
	"app/codeAssistant/internal/httphandler"
	"app/codeAssistant/internal/lm_settings"
	"app/codeAssistant/internal/scheduling"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// ============================================================
	// 1. Инициализация БД (Лучше делать это ПЕРВЫМ)
	// ============================================================

	// Создаем папку data, если нет
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0755); err != nil {
			log.Fatalf("❌ Не удалось создать папку  %v", err)
		}
	}

	dbHandler := &db.DbHandler{}
	if err := dbHandler.InitDB("data/tasks.db"); err != nil {
		log.Fatalf("❌ Ошибка БД: %v", err)
	}
	if dbHandler.Db == nil {
		log.Fatal("❌ Критическая ошибка: подключение к БД равно nil")
	}
	defer dbHandler.Db.Close()
	log.Println("✅ База данных подключена")

	// ============================================================
	// 2. Инициализация ИИ и запуск в фоне
	// ============================================================

	kasper := lm_settings.NewLMStudioClient("http://127.0.0.1:1234", "", "qwen2.5-coder-7b-instruct")
	sysPrompt := scheduling.NewSystemPrompt()

	basePrompt := scheduling.NewBasePromptSet(
		"Напиши функцию которая будет рисовать елку в .txt файле. Используй язык golang",
		[]scheduling.Message{},
		scheduling.AgentState{},
		"normal",
	)

	// Создаем итератор
	iter := executor.NewIterator(*kasper, *sysPrompt, *basePrompt)

	// 🚀 ЗАПУСКАЕМ В ГОРУТИНЕ (ключевое слово go)
	// Теперь ИИ работает в фоне, а код идет дальше сразу же
	go func() {
		log.Println("🤖 Запуск ИИ-агента...")
		resp, err := iter.StartIteration()
		if err != nil {
			log.Printf("⚠️ Ошибка агента: %v", err)
			return
		}
		log.Println("✅ Агент завершил работу.")
		fmt.Println("📄 Результат кода:")
		fmt.Println(resp.Parameters.Code)

		// Здесь можно добавить логику сохранения результата в БД, если нужно
		// Например: dbHandler.CreateTask("Результат ИИ", resp.Parameters.Code, scheduling.High)
	}()

	// ============================================================
	// 3. Запуск HTTP Сервера
	// ============================================================
	server := httphandler.NewServer(dbHandler)

	http.HandleFunc("/", server.IndexHandler)
	http.HandleFunc("/add", server.AddHandler)
	http.HandleFunc("/edit", server.EditHandler)
	http.HandleFunc("/delete", server.DeleteHandler)
	http.HandleFunc("/reset", server.ResetHandler)
	http.HandleFunc("/api/update-status", server.UpdateStatusAPIHandler)
	http.HandleFunc("/api/reorder-tasks", server.ReorderTasksHandler)
	http.HandleFunc("/api/tasks", server.TaskAPIHandler)

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Println("📂 Откройте браузер, пока агент работает в фоне...")

	// Запуск сервера (блокирует основной поток, но это и нужно)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
