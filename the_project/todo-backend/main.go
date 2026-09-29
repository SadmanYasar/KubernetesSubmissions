package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

var db *sql.DB

func initDB() (*sql.DB, error) {
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "postgres-svc"
	}
	dbPort := os.Getenv("POSTGRES_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgrespassword"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "todos"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to Postgres database at %s:%s (db=%s)...", dbHost, dbPort, dbName)

	var database *sql.DB
	var err error

	// Retry connecting for up to 60 seconds to allow Postgres pod to start
	for attempts := 1; attempts <= 30; attempts++ {
		database, err = sql.Open("postgres", connStr)
		if err == nil {
			err = database.Ping()
			if err == nil {
				log.Println("Connected to Postgres successfully!")
				break
			}
		}
		log.Printf("Waiting for database connection (attempt %d/30): %v", attempts, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		text VARCHAR(140) NOT NULL
	);
	`
	_, err = database.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}

	// Seed default todos if table is empty
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count); err == nil && count == 0 {
		initialTodos := []string{
			"Read DevOps with Kubernetes course materials",
			"Deploy app with local PersistentVolumes",
			"Share data between Ping-pong and Log-output",
		}
		for _, text := range initialTodos {
			_, _ = database.Exec("INSERT INTO todos (text) VALUES ($1)", text)
		}
		log.Println("Seeded initial todos into database.")
	}

	log.Println("Database table todos is ready.")
	return database, nil
}

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
	rows, err := db.Query("SELECT id, text FROM todos ORDER BY id ASC")
	if err != nil {
		log.Printf("Failed to query todos: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text); err != nil {
			log.Printf("Failed to scan todo row: %v", err)
			continue
		}
		result = append(result, t)
	}

	if result == nil {
		result = []Todo{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
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

	var newTodo Todo
	err = db.QueryRow("INSERT INTO todos (text) VALUES ($1) RETURNING id, text", text).Scan(&newTodo.ID, &newTodo.Text)
	if err != nil {
		log.Printf("Failed to insert todo: %v", err)
		http.Error(w, "Failed to save todo to database", http.StatusInternalServerError)
		return
	}

	log.Printf("Created new todo in database: [ID: %d] %s", newTodo.ID, newTodo.Text)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if db != nil {
		if err := db.Ping(); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Fatal: could not initialize database: %v", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/healthz", healthHandler)

	fmt.Printf("Todo-backend service started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
