package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	counter  int
	mu       sync.Mutex
	filePath = "/usr/src/app/shared/pongs.txt"
)

func initCounter() {
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err == nil {
			val, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err == nil {
				counter = val
				log.Printf("Initialized counter from file: %d", counter)
				return
			}
		}
	}
	counter = 0
	log.Printf("Initialized counter to 0")
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(w, "pong %d", counter)
	counter++

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create directory %s: %v", dir, err)
		return
	}

	// Write the new counter value to file
	err := os.WriteFile(filePath, []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		log.Printf("Failed to write counter to file: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	initCounter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/pingpong", pingPongHandler)
	http.HandleFunc("/", rootHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

