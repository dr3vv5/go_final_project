package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dr3vv5/go_final_project/pkg/api"
	"github.com/dr3vv5/go_final_project/pkg/db"
	"github.com/dr3vv5/go_final_project/pkg/models"
	"github.com/go-chi/chi/v5"
)

type CreateTaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type UpdateTaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	// 1. Сначала ставим заголовок и статус код!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 2. Потом пишем тело
	response := map[string]string{
		"error": message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Если даже записать ошибку не удалось (клиент отключился), мы ничего не можем сделать.
		// Но главное — статус код уже отправлен выше.
		return
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	search := r.URL.Query().Get("search")

	var dateFilter string
	var isDateSearch bool

	if search != "" {
		parsedDate, err := time.Parse("02.01.2006", search)
		if err == nil {
			isDateSearch = true
			dateFilter = parsedDate.Format("20060102")
		}
	}

	var tasks []models.Task
	var err error

	if isDateSearch {
		tasks, err = storage.GetTasksByDate(dateFilter)
	} else {
		tasks, err = storage.GetTasksBySearch(search)
	}

	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responseTasks = []TaskResponse{}
	for _, t := range tasks {
		idStr := strconv.Itoa(t.ID)
		dateStr := t.Date.Format("20060102")
		responseTasks = append(responseTasks, TaskResponse{
			ID:      idStr,
			Date:    dateStr,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})
	}

	response := map[string]any{
		"tasks": responseTasks,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func CreateTask(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		writeJSONError(w, "Field 'title' is required", http.StatusBadRequest)
		return
	}

	finalDate, err := resolveDate(req.Date, req.Repeat)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	newTask := models.Task{
		Title:   req.Title,
		Comment: req.Comment,
		Date:    finalDate,
		Repeat:  req.Repeat,
	}

	id, err := storage.CreateTask(newTask)
	if err != nil {
		writeJSONError(w, "Failed to save task to database", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"id": strconv.Itoa(id),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func resolveDate(dateStr, repeatStr string) (time.Time, error) {
	now := time.Now()

	if dateStr == "" {
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}

	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("некорректный формат даты")
	}

	if !api.AfterNow(t, now) {
		if repeatStr == "" {
			return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
		} else {
			nextDateStr, err := api.NextDate(now, dateStr, repeatStr)
			if err != nil {
				return time.Time{}, fmt.Errorf("ошибка в правиле повторения: %w", err)
			}

			finalDate, err := time.Parse("20060102", nextDateStr)
			if err != nil {
				return time.Time{}, fmt.Errorf("ошибка парсинга вычисленной даты: %w", err)
			}
			return finalDate, nil
		}
	}

	return t, nil
}
func SetupRoutes(r chi.Router, storage *db.Storage) {
	r.Post("/signin", SigninHandler)

	r.Get("/tasks", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		GetTasks(w, r, storage)
	}))

	r.Get("/task", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		GetTaskHandler(w, r, storage)
	}))

	r.Post("/task", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		CreateTask(w, r, storage)
	}))

	r.Get("/nextdate", NextDateHandler)

	r.Put("/task", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		UpdateTaskHandler(w, r, storage)
	}))

	r.Delete("/task", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		DeleteTaskHandler(w, r, storage)
	}))

	r.Post("/task/done", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		MarkTaskAsDoneHandler(w, r, storage)
	}))
}
