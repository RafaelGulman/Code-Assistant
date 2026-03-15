package httphandler

import (
	"app/codeAssistant/internal/scheduling"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tasks, err := s.DB.GetTasks()
	if err != nil {
		http.Error(w, "Ошибка получения задач: "+err.Error(), http.StatusInternalServerError)
		log.Printf("[ERROR] GetTasks: %v", err)
		return
	}

	// Передаём задачи в шаблон
	err = s.Templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Tasks": tasks,
		"StatusNames": map[scheduling.TaskStatus]string{
			scheduling.Stop:    "Остановлено",
			scheduling.Process: "В процессе",
			scheduling.Finish:  "Завершено",
			scheduling.Error:   "Ошибка",
		},
		"PriorityNames": map[scheduling.TaskPriority]string{
			scheduling.Low:    "Низкий",
			scheduling.Medium: "Средний",
			scheduling.High:   "Высокий",
		},
	})
	if err != nil {
		http.Error(w, "Ошибка рендеринга: "+err.Error(), http.StatusInternalServerError)
	}
}

// AddHandler — добавление новой задачи
func (s *Server) AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Показываем форму
		err := s.Templates.ExecuteTemplate(w, "form.html", map[string]interface{}{
			"IsEdit": false,
			"Task":   scheduling.Task{},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Парсим форму
		_ = r.ParseForm()
		title := r.FormValue("title")
		description := r.FormValue("description")
		priority, _ := strconv.Atoi(r.FormValue("priority"))

		err := s.DB.CreateTask(title, description, scheduling.TaskPriority(priority))
		if err != nil {
			http.Error(w, "Ошибка создания: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// EditHandler — редактирование задачи
func (s *Server) EditHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		task, err := s.DB.GetTaskByID(id)
		if err != nil {
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}

		err = s.Templates.ExecuteTemplate(w, "form.html", map[string]interface{}{
			"IsEdit": true,
			"Task":   task,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		title := r.FormValue("title")
		description := r.FormValue("description")
		priority, _ := strconv.Atoi(r.FormValue("priority"))
		status, _ := strconv.Atoi(r.FormValue("status"))

		err := s.DB.UpdateTask(id, title, description,
			scheduling.TaskPriority(priority),
			scheduling.TaskStatus(status))
		if err != nil {
			http.Error(w, "Ошибка обновления: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// DeleteHandler — удаление задачи
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	err = s.DB.DeleteTask(id)
	if err != nil {
		http.Error(w, "Ошибка удаления: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ResetHandler — сброс БД (удаление таблицы)
func (s *Server) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := s.DB.DropTasksTable()
	if err != nil {
		http.Error(w, "Ошибка сброса: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Пересоздаём таблицу
	_ = s.DB.InitDB(s.DB.DbName)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ==================== API (AJAX) ====================

// UpdateStatusAPIHandler — AJAX-эндпоинт для изменения статуса
func (s *Server) UpdateStatusAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_ = r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))
	status, _ := strconv.Atoi(r.FormValue("status"))

	err := s.DB.UpdateTaskStatus(id, scheduling.TaskStatus(status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем актуальные времена для обновления карточки
	task, _ := s.DB.GetTaskByID(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"change_status": task.ChangeStatus.Format("02.01 15:04"),
		"begin_task":    task.BeginTask.Format("02.01 15:04"),
		"end_task": func() string {
			if task.EndTask.IsZero() {
				return ""
			}
			return task.EndTask.Format("02.01 15:04")
		}(),
	})
}

// TaskAPIHandler — REST API для получения задач (для ИИ-агента)
func (s *Server) TaskAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET /api/tasks — получить все задачи
	// GET /api/tasks?status=1 — получить задачи со статусом "В процессе"

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := s.DB.GetTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Фильтрация по статусу (опционально)
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" {
		filterStatus, _ := strconv.Atoi(statusFilter)
		filtered := make([]scheduling.Task, 0)
		for _, t := range tasks {
			if int(t.Status) == filterStatus {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	json.NewEncoder(w).Encode(tasks)
}
