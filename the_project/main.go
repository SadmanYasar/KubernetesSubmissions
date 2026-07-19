package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	imagePath = "/app/shared/image.jpg"
	mu        sync.Mutex
)

func downloadImage(path string) error {
	resp, err := http.Get("https://picsum.photos/1200")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	info, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		log.Println("Image cache not found. Downloading image...")
		if err := downloadImage(imagePath); err != nil {
			log.Printf("Failed to download image: %v", err)
			http.Error(w, "Failed to download image", http.StatusInternalServerError)
			return
		}
		info, err = os.Stat(imagePath)
		if err != nil {
			http.Error(w, "Failed to read downloaded image info", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Error inspecting image cache", http.StatusInternalServerError)
		return
	}

	// Serve the cached image
	http.ServeFile(w, r, imagePath)

	// If the file is older than 10 minutes, trigger background download for next request
	if time.Since(info.ModTime()) > 10*time.Minute {
		log.Println("Image cache is older than 10 minutes. Triggering background update...")
		go func() {
			tempPath := imagePath + ".tmp"
			if err := downloadImage(tempPath); err == nil {
				mu.Lock()
				os.Rename(tempPath, imagePath)
				mu.Unlock()
				log.Println("Image cache updated successfully in background.")
			} else {
				log.Printf("Background image download failed: %v", err)
				os.Remove(tempPath)
			}
		}()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>The Project - Todos</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				background-color: #f0f2f5;
				margin: 0;
				padding: 20px;
				min-height: 100vh;
			}
			.container {
				background: white;
				padding: 24px;
				border-radius: 8px;
				box-shadow: 0 4px 6px rgba(0,0,0,0.1);
				text-align: center;
				max-width: 600px;
				width: 100%;
			}
			img {
				max-width: 100%;
				height: auto;
				border-radius: 4px;
				margin-top: 16px;
				margin-bottom: 24px;
			}
			.todo-section {
				margin-top: 24px;
				text-align: left;
			}
			.todo-form {
				display: flex;
				gap: 8px;
				margin-bottom: 16px;
			}
			.todo-input {
				flex-grow: 1;
				padding: 8px 12px;
				border: 1px solid #ccc;
				border-radius: 4px;
				font-size: 14px;
			}
			.todo-btn {
				background-color: #007bff;
				color: white;
				border: none;
				padding: 8px 16px;
				border-radius: 4px;
				cursor: pointer;
				font-size: 14px;
				font-weight: bold;
			}
			.todo-btn:hover {
				background-color: #0056b3;
			}
			ul {
				list-style-type: none;
				padding: 0;
				margin: 0;
			}
			li {
				background: #f8f9fa;
				padding: 10px 14px;
				border-bottom: 1px solid #dee2e6;
				border-radius: 4px;
				margin-bottom: 8px;
				font-size: 14px;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Hello from Kubernetes!</h1>
			<p>This is a simple HTML page served by the project.</p>
			<img src="/image.jpg" alt="Hourly Random Image" />
			
			<div class="todo-section">
				<h3>Todo List</h3>
				<form class="todo-form" onsubmit="event.preventDefault(); alert('Todo creation will be implemented in a future exercise!');">
					<input class="todo-input" type="text" maxlength="140" placeholder="Enter todo (max 140 chars)..." required />
					<button class="todo-btn" type="submit">Send</button>
				</form>
				<ul>
					<li>📝 Read DevOps with Kubernetes course materials</li>
					<li>⚙️ Deploy app with local PersistentVolumes</li>
					<li>🏓 Share data between Ping-pong and Log-output</li>
				</ul>
			</div>
		</div>
	</body>
	</html>
	`
	fmt.Fprint(w, html)
}

func crashHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Crash endpoint hit! Shutting down container...")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Crashing container... Check Kubernetes to see the pod restart!"))

	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Starting server in port %s\n", port)

	http.HandleFunc("/", handler)
	http.HandleFunc("/image.jpg", imageHandler)
	http.HandleFunc("/crash", crashHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

