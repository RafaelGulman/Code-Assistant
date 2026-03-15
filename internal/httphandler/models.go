package httphandler

import (
	"app/codeAssistant/internal/db"
	"html/template"
)

// Server — HTTP-сервер с зависимостями
type Server struct {
	DB        *db.DbHandler
	Templates *template.Template
}

// NewServer — конструктор сервера
func NewServer(dbHandler *db.DbHandler) *Server {
	// Парсим шаблоны при старте (в продакшене можно кэшировать)
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	return &Server{
		DB:        dbHandler,
		Templates: tmpl,
	}
}
