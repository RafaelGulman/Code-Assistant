package db

import (
	"app/codeAssistant/internal/scheduling"
	"database/sql"
	"fmt"
	"log"
	"time"
)

func (h *DbHandler) InitDB(dbPath string) error {
	var err error

	// 1. Попытка открытия
	h.Db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		// ❌ БЫЛО: return nil (это скрывало проблему)
		// ✅ СТАЛО: Возвращаем реальную ошибку
		return fmt.Errorf("ошибка sql.Open: %w", err)
	}

	// 2. Проверка реального соединения (Ping)
	// sql.Open может вернуть успех, даже если файл недоступен. Ping проверяет реально.
	if err := h.Db.Ping(); err != nil {
		h.Db.Close() // Закрываем битое соединение
		h.Db = nil   // Сбрасываем в nil, чтобы было понятно
		return fmt.Errorf("ошибка подключения (Ping): %w", err)
	}

	// 3. Создание таблицы
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		priority INTEGER DEFAULT 1,
		status INTEGER DEFAULT 0,
		begin_task DATETIME,
		end_task DATETIME,
		change_status DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = h.Db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	log.Println("✅ База данных успешно инициализирована:", dbPath)
	return nil
}

// CreateTask — добавление новой задачи
func (db *DbHandler) CreateTask(title, description string, priority scheduling.TaskPriority) error {
	insertSQL := `
	INSERT INTO tasks (title, description, priority, status, change_status) 
	VALUES (?, ?, ?, ?, ?)`
	_, err := db.Db.Exec(insertSQL, title, description, priority, scheduling.Stop, time.Now())
	return err
}

// GetTasks — получение всех задач
func (db *DbHandler) GetTasks() ([]scheduling.Task, error) {
	selectSQL := `
	SELECT id, title, description, priority, status, 
		   begin_task, end_task, change_status 
	FROM tasks ORDER BY change_status DESC`

	rows, err := db.Db.Query(selectSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []scheduling.Task
	for rows.Next() {
		var t scheduling.Task
		var beginTask, endTask, changeStatus sql.NullString

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Priority, &t.Status,
			&beginTask, &endTask, &changeStatus,
		)
		if err != nil {
			return nil, err
		}

		// Обработка nullable datetime полей
		if beginTask.Valid {
			t.BeginTask, _ = time.Parse("2006-01-02 15:04:05", beginTask.String)
		}
		if endTask.Valid {
			t.EndTask, _ = time.Parse("2006-01-02 15:04:05", endTask.String)
		}
		if changeStatus.Valid {
			t.ChangeStatus, _ = time.Parse("2006-01-02 15:04:05", changeStatus.String)
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetTaskByID — получение задачи по ID
func (db *DbHandler) GetTaskByID(id int) (*scheduling.Task, error) {
	selectSQL := `
	SELECT id, title, description, priority, status, 
		   begin_task, end_task, change_status 
	FROM tasks WHERE id = ?`

	row := db.Db.QueryRow(selectSQL, id)
	var t scheduling.Task
	var beginTask, endTask, changeStatus sql.NullString

	err := row.Scan(
		&t.ID, &t.Title, &t.Description, &t.Priority, &t.Status,
		&beginTask, &endTask, &changeStatus,
	)
	if err != nil {
		return nil, err
	}

	if beginTask.Valid {
		t.BeginTask, _ = time.Parse("2006-01-02 15:04:05", beginTask.String)
	}
	if endTask.Valid {
		t.EndTask, _ = time.Parse("2006-01-02 15:04:05", endTask.String)
	}
	if changeStatus.Valid {
		t.ChangeStatus, _ = time.Parse("2006-01-02 15:04:05", changeStatus.String)
	}

	return &t, nil
}

// UpdateTask — полное обновление задачи
func (db *DbHandler) UpdateTask(id int, title, description string, priority scheduling.TaskPriority, status scheduling.TaskStatus) error {
	updateSQL := `
	UPDATE tasks 
	SET title = ?, description = ?, priority = ?, status = ?, change_status = ? 
	WHERE id = ?`
	_, err := db.Db.Exec(updateSQL, title, description, priority, status, time.Now(), id)
	return err
}

// UpdateTaskStatus — точечное обновление только статуса (с таймстемпом)
func (db *DbHandler) UpdateTaskStatus(id int, status scheduling.TaskStatus) error {
	now := time.Now()
	var updateSQL string
	var err error

	switch status {
	case scheduling.Process:
		// При переходе в "В процессе" фиксируем время начала
		updateSQL = `UPDATE tasks SET status = ?, change_status = ?, begin_task = COALESCE(begin_task, ?) WHERE id = ?`
		_, err = db.Db.Exec(updateSQL, status, now, now, id)
	case scheduling.Finish, scheduling.Error, scheduling.Stop:
		// При завершении/ошибке/остановке фиксируем время окончания
		updateSQL = `UPDATE tasks SET status = ?, change_status = ?, end_task = ? WHERE id = ?`
		_, err = db.Db.Exec(updateSQL, status, now, now, id)
	default:
		updateSQL = `UPDATE tasks SET status = ?, change_status = ? WHERE id = ?`
		_, err = db.Db.Exec(updateSQL, status, now, id)
	}
	return err
}

// DeleteTask — удаление задачи
func (db *DbHandler) DeleteTask(id int) error {
	deleteSQL := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Db.Exec(deleteSQL, id)
	return err
}

// DropTasksTable — удаление всей таблицы (сброс)
func (db *DbHandler) DropTasksTable() error {
	dropSQL := `DROP TABLE IF EXISTS tasks`
	_, err := db.Db.Exec(dropSQL)
	return err
}

// UpdateTasksOrder обновляет приоритеты задач согласно новому списку ID
func (db *DbHandler) UpdateTasksOrder(ids []int) error {
	// Начинаем транзакцию для скорости и надежности
	tx, err := db.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE tasks SET priority = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, id := range ids {
		var priority int
		// Присваиваем приоритет в зависимости от позиции
		if i == 0 {
			priority = 2 // High
		} else if i == 1 {
			priority = 1 // Medium
		} else {
			priority = 0 // Low
		}

		_, err := stmt.Exec(priority, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
