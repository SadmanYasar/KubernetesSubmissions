package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	counter int
	mu      sync.Mutex
)

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(w, "pong %d", counter)
	counter++
}

func pingsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(w, "%d", counter)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
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

