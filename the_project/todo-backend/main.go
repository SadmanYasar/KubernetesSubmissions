package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

var (
	todos = []Todo{
		{ID: 1, Text: "Read DevOps with Kubernetes course materials"},
		{ID: 2, Text: "Deploy app with local PersistentVolumes"},
		{ID: 3, Text: "Share data between Ping-pong and Log-output"},
	}
	nextID = 4
	mu     sync.RWMutex
)

func enableCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTodos(w, r)
	case http.MethodPost:
		createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		log.Printf("Failed to encode todos: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type CreateTodoRequest struct {
	Text string `json:"text"`
	Todo string `json:"todo"` // Fallback alias
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var text string
	var req CreateTodoRequest
	if err := json.Unmarshal(body, &req); err == nil {
		if req.Text != "" {
			text = req.Text
		} else if req.Todo != "" {
			text = req.Todo
		}
	} else {
		// Fallback to plain text if body is not JSON
		text = strings.TrimSpace(string(body))
	}

	text = strings.TrimSpace(text)
	if text == "" {
		http.Error(w, "Todo text cannot be empty", http.StatusBadRequest)
		return
	}

	if len(text) > 140 {
		http.Error(w, "Todo text cannot exceed 140 characters", http.StatusBadRequest)
		return
	}

	mu.Lock()
	newTodo := Todo{
		ID:   nextID,
		Text: text,
	}
	nextID++
	todos = append(todos, newTodo)
	mu.Unlock()

	log.Printf("Created new todo: [ID: %d] %s", newTodo.ID, newTodo.Text)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/healthz", healthHandler)

	fmt.Printf("Todo-backend service started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
