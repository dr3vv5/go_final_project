package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dr3vv5/go_final_project/pkg/db"
	"github.com/dr3vv5/go_final_project/pkg/server"
)

func main() {

	const defaultDBFile = "scheduler.db"
	const envVarName = "TODO_DBFILE"

	dbFile := os.Getenv(envVarName)
	if dbFile == "" {
		dbFile = defaultDBFile
		log.Printf("Переменная окружения %s не задана. Используем путь по умолчанию: %s", envVarName, dbFile)
	} else {
		log.Printf("Обнаружена переменная %s = %s. Используем указанный путь.", envVarName, dbFile)
	}

	// 2. Превращаем в абсолютный путь (наша страховка)
	absPath, err := filepath.Abs(dbFile)
	if err != nil {
		log.Fatalf("Не удалось получить абсолютный путь к БД: %v", err)
	}

	log.Printf("Инициализация БД по пути: %s", absPath)

	storage, err := db.NewStorage(absPath)
	if err != nil {
		log.Fatalf("Критическая ошибка инициализации БД: %v", err)
	}

	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения с БД: %v", err)
		}
		log.Println("Соединение с БД закрыто.")
	}()

	port := getPortFromEnv()

	log.Printf("Starting server on port %d...\n", port)

	if err := server.Run(port, storage); err != nil {
		log.Fatalf("Critical error while starting the server: %v", err)
	}
}

func getPortFromEnv() int {
	const defaultPort = 7540
	const envVarName = "TODO_PORT"

	portStr := os.Getenv(envVarName)

	if portStr == "" {
		return defaultPort
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Warning: Invalid port format in %s ('%s'). Using default port %d. Error: %v",
			envVarName, portStr, defaultPort, err)
		return defaultPort
	}

	if portInt <= 0 {
		log.Printf("Warning: Port %d is invalid (must be greater than 0). Using default port %d.",
			portInt, defaultPort)
		return defaultPort
	}

	return portInt
}
