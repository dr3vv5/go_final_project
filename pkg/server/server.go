package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dr3vv5/go_final_project/pkg/db"
	"github.com/dr3vv5/go_final_project/pkg/handlers"
	"github.com/go-chi/chi/v5"
)

func Run(port int, storage *db.Storage) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get caller info")
	}
	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "..", "..")
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve project root: %w", err)
	}

	webDir := filepath.Join(projectRoot, "web")

	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		log.Printf("Static files directory not found at '%s'. Serving only API.", webDir)
	} else {
		log.Printf("Serving static files from: %s", webDir)
	}

	r := chi.NewRouter()

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			return
		}
		filePath := filepath.Join(webDir, r.URL.Path)
		http.ServeFile(w, r, filePath)
	})

	r.Route("/api", func(r chi.Router) {
		handlers.SetupRoutes(r, storage)
	})

	return http.ListenAndServe(addr, r)
}
