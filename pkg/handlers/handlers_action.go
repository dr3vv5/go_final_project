package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dr3vv5/go_final_project/pkg/api"
	"github.com/dr3vv5/go_final_project/pkg/db"
	"github.com/dr3vv5/go_final_project/pkg/models"
)

func GetTaskHandler(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	// 1. Достаем ID из query-параметров (?id=123), а не из пути
	idStr := r.URL.Query().Get("id")

	// 2. Если ID вообще не передан, это ошибка клиента
	if idStr == "" {
		writeJSONError(w, "Parameter 'id' is required", http.StatusBadRequest)
		return
	}

	// 3. Валидируем, что ID — это число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Invalid ID format. ID must be an integer.", http.StatusBadRequest)
		return
	}

	// 4. Обращаемся к базе данных
	task, err := storage.GetTask(id)

	// 5. Обрабатываем ошибки
	if err != nil {
		if err == sql.ErrNoRows {
			// Задача не найдена -> 404
			writeJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			// Ошибка БД -> 500
			writeJSONError(w, "Failed to retrieve task from database", http.StatusInternalServerError)
		}
		return
	}

	// 6. Формируем успешный ответ
	response := TaskResponse{
		ID:      strconv.Itoa(task.ID),
		Date:    task.Date.Format("20060102"),
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	var rawData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	idRaw, ok := rawData["id"]
	if !ok {
		writeJSONError(w, "Field 'id' is required", http.StatusBadRequest)
		return
	}

	idStr, okStr := idRaw.(string)
	if !okStr {
		if idNum, okNum := idRaw.(float64); okNum {
			idStr = strconv.FormatInt(int64(idNum), 10)
		} else {
			writeJSONError(w, "Field 'id' must be a valid number", http.StatusBadRequest)
			return
		}
	}

	if idStr == "" {
		writeJSONError(w, "Field 'id' cannot be empty", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Invalid ID format. ID must be an integer.", http.StatusBadRequest)
		return
	}

	dateStr, _ := rawData["date"].(string)
	titleStr, _ := rawData["title"].(string)
	commentStr, _ := rawData["comment"].(string)
	repeatStr, _ := rawData["repeat"].(string)

	if titleStr == "" {
		writeJSONError(w, "Field 'title' is required", http.StatusBadRequest)
		return
	}

	finalDate, err := resolveDate(dateStr, repeatStr)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedTask := models.Task{
		ID:      id,
		Title:   titleStr,
		Comment: commentStr,
		Date:    finalDate,
		Repeat:  repeatStr,
	}

	err = storage.UpdateTask(updatedTask)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			writeJSONError(w, "Failed to update task", http.StatusInternalServerError)
		}
		return
	}

	response := TaskResponse{
		ID:      strconv.Itoa(updatedTask.ID),
		Date:    updatedTask.Date.Format("20060102"),
		Title:   updatedTask.Title,
		Comment: updatedTask.Comment,
		Repeat:  updatedTask.Repeat,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	if dateStr == "" || repeatStr == "" {
		http.Error(w, "Ошибка: параметры 'date' и 'repeat' обязательны", http.StatusBadRequest)
		return
	}

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка формата даты 'now': %v", err), http.StatusBadRequest)
			return
		}
	}

	result, err := api.NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, result)
}
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	// 1. Получаем ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSONError(w, "Parameter 'id' is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Invalid ID format. ID must be an integer.", http.StatusBadRequest)
		return
	}

	err = storage.DeleteTask(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			writeJSONError(w, "Failed to delete task from database", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {

		return
	}
}
func MarkTaskAsDoneHandler(w http.ResponseWriter, r *http.Request, storage *db.Storage) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSONError(w, "Parameter 'id' is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "Invalid ID format. ID must be an integer.", http.StatusBadRequest)
		return
	}

	task, err := storage.GetTask(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			writeJSONError(w, "Failed to retrieve task", http.StatusInternalServerError)
		}
		return
	}

	if task.Repeat == "" {
		err = storage.DeleteTask(id)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSONError(w, "Task was deleted concurrently", http.StatusNotFound)
			} else {
				writeJSONError(w, "Failed to delete task", http.StatusInternalServerError)
			}
			return
		}
	} else {
		currentDateStr := task.Date.Format("20060102")
		nextDateStr, err := api.NextDate(time.Now(), currentDateStr, task.Repeat)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Invalid repeat rule: %v", err), http.StatusBadRequest)
			return
		}

		newDate, err := time.Parse("20060102", nextDateStr)
		if err != nil {
			writeJSONError(w, "Failed to parse calculated date", http.StatusInternalServerError)
			return
		}

		err = storage.UpdateTaskDate(id, newDate)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSONError(w, "Task was deleted concurrently", http.StatusNotFound)
			} else {
				writeJSONError(w, "Failed to update task date", http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		return
	}
}
