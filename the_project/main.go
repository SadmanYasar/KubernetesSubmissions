//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Hello from Kubernetes!</h1><p>This is a simple HTML page served by the project.</p>")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port if not specified
	}
	fmt.Printf("Starting server in port %s\n", port)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
