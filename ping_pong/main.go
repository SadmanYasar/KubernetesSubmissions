package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

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
		dbName = "postgres"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to Postgres database at %s:%s (db=%s)...", dbHost, dbPort, dbName)

	var database *sql.DB
	var err error

	// Retry loop to wait for postgres to be ready
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
	CREATE TABLE IF NOT EXISTS pingpong (
		id INT PRIMARY KEY,
		counter INT NOT NULL
	);
	INSERT INTO pingpong (id, counter)
	VALUES (1, 0)
	ON CONFLICT (id) DO NOTHING;
	`
	_, err = database.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}

	log.Println("Database table pingpong initialized.")
	return database, nil
}

func getCounter() (int, error) {
	var count int
	err := db.QueryRow("SELECT counter FROM pingpong WHERE id = 1").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func incrementCounter() (int, error) {
	var newCount int
	// We increment the counter in the DB and return the value prior to increment (or current count)
	// Notice exercise: first visit -> pong 0, counter becomes 1.
	err := db.QueryRow(`
		UPDATE pingpong
		SET counter = counter + 1
		WHERE id = 1
		RETURNING counter - 1
	`).Scan(&newCount)
	if err != nil {
		return 0, err
	}
	return newCount, nil
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	currentVal, err := incrementCounter()
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "pong %d", currentVal)
}

func pingsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := getCounter()
	if err != nil {
		log.Printf("Error getting counter: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", count)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
		port = "8080"
	}

	http.HandleFunc("/pingpong", pingPongHandler)
	http.HandleFunc("/pings", pingsHandler)
	http.HandleFunc("/", rootHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
