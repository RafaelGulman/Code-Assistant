package db

import (
	"app/codeAssistant/internal/scheduling"
	"database/sql"
	"time"
)

// InitDB — инициализация подключения и создание таблицы
func (db *DbHandler) InitDB(dbPath string) error {
	var err error
	db.DbName = dbPath
	db.Db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil
	}

	// SQL запрос на создание таблицы
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

	_, err = db.Db.Exec(createTableSQL)
	if err != nil {
		return nil
	}

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
