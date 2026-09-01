package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/dr3vv5/go_final_project/pkg/models"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	dataSourceName := fmt.Sprintf("file:%s?mode=rwc", dbPath)

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение с БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка проверки соединения: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка применения схемы БД: %w", err)
	}

	log.Printf("База данных успешно инициализирована по пути: %s", dbPath)

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255),
    comment TEXT,
    date CHAR(8),
    repeat VARCHAR(128)
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

func (s *Storage) CreateTask(task models.Task) (int, error) {
	dateString := task.Date.Format("20060102")

	query := `INSERT INTO scheduler (title, comment, date, repeat) VALUES (?, ?, ?, ?)`

	res, err := s.db.Exec(query, task.Title, task.Comment, dateString, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *Storage) GetAllTasks() ([]models.Task, error) {
	query := `SELECT id, title, comment, date, repeat FROM scheduler ORDER BY date LIMIT 50`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении задач: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dateStr string
		err := rows.Scan(&t.ID, &t.Title, &t.Comment, &dateStr, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		parsedDate, err := time.Parse("20060102", dateStr)
		if err != nil {
			return nil, fmt.Errorf("некорректный формат даты в БД: %s, ошибка: %w", dateStr, err)
		}
		t.Date = parsedDate
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return tasks, nil
}
func (s *Storage) GetTasksByDate(dateStr string) ([]models.Task, error) {
	query := `SELECT id, title, comment, date, repeat FROM scheduler 
               WHERE date = ? 
               ORDER BY date 
               LIMIT 50`

	rows, err := s.db.Query(query, dateStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске по дате: %w", err)
	}
	defer rows.Close()

	return s.scanTasks(rows)
}

func (s *Storage) GetTasksBySearch(search string) ([]models.Task, error) {
	if search == "" {
		return s.GetAllTasks()
	}

	pattern := "%" + search + "%"

	query := `SELECT id, title, comment, date, repeat FROM scheduler 
               WHERE title LIKE ? COLLATE NOCASE 
                  OR comment LIKE ? COLLATE NOCASE 
               ORDER BY date 
               LIMIT 50`

	rows, err := s.db.Query(query, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске по тексту: %w", err)
	}
	defer rows.Close()

	return s.scanTasks(rows)
}

func (s *Storage) scanTasks(rows *sql.Rows) ([]models.Task, error) {
	var tasks []models.Task

	for rows.Next() {
		var t models.Task
		var dateStr string

		if err := rows.Scan(&t.ID, &t.Title, &t.Comment, &dateStr, &t.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}

		parsedDate, err := time.Parse("20060102", dateStr)
		if err != nil {
			return nil, fmt.Errorf("некорректный формат даты в БД: %s, ошибка: %w", dateStr, err)
		}
		t.Date = parsedDate

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return tasks, nil
}
func (s *Storage) GetTask(id int) (models.Task, error) {
	var task models.Task
	var dateStr string

	query := "SELECT id, title, comment, date, repeat FROM scheduler WHERE id = ?"
	row := s.db.QueryRow(query, id)
	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Comment,
		&dateStr,
		&task.Repeat,
	)
	if err != nil {
		return task, err
	}
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return task, fmt.Errorf("некорректный формат даты в БД: %s, ошибка: %w", dateStr, err)
	}
	task.Date = parsedDate
	return task, nil
}
func (s *Storage) UpdateTask(task models.Task) error {
	dateStr := task.Date.Format("20060102")

	res, err := s.db.Exec(
		`UPDATE scheduler SET title = ?, comment = ?, date = ?, repeat = ? WHERE id = ?`,
		task.Title, task.Comment, dateStr, task.Repeat, task.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// ГЛАВНОЕ: Если обновилось 0 строк — значит, задачи с таким ID не было!
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *Storage) DeleteTask(id int) error {
	// Используем тот же плейсхолдер ?, что и в CreateTask/UpdateTask
	query := `DELETE FROM scheduler WHERE id = ?`

	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении задачи: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	// Ключевой момент: если rowsAffected == 0, значит, задачи не было.
	// Возвращаем sql.ErrNoRows, чтобы хендлер мог вернуть 404.
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *Storage) UpdateTaskDate(id int, newDate time.Time) error {
	dateStr := newDate.Format("20060102")

	query := `UPDATE scheduler SET date = ? WHERE id = ?`

	res, err := s.db.Exec(query, dateStr, id)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении даты: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
