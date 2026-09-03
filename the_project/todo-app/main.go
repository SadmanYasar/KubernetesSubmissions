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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>The Project - Todos</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			background-color: #f0f2f5;
			margin: 0;
			padding: 20px;
			min-height: 100vh;
			box-sizing: border-box;
		}
		.container {
			background: white;
			padding: 30px;
			border-radius: 12px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.08);
			text-align: center;
			max-width: 600px;
			width: 100%;
			box-sizing: border-box;
		}
		h1 {
			margin-top: 0;
			color: #1a1a1a;
			font-size: 24px;
		}
		p {
			color: #666;
			margin-bottom: 20px;
		}
		img {
			max-width: 100%;
			height: auto;
			border-radius: 8px;
			margin-bottom: 24px;
			box-shadow: 0 2px 6px rgba(0,0,0,0.06);
		}
		.todo-section {
			margin-top: 10px;
			text-align: left;
		}
		.todo-section h3 {
			margin-bottom: 12px;
			color: #333;
		}
		.todo-form {
			display: flex;
			gap: 10px;
			margin-bottom: 8px;
		}
		.todo-input {
			flex-grow: 1;
			padding: 10px 14px;
			border: 1px solid #d1d5db;
			border-radius: 6px;
			font-size: 14px;
			outline: none;
			transition: border-color 0.2s;
		}
		.todo-input:focus {
			border-color: #007bff;
		}
		.todo-btn {
			background-color: #007bff;
			color: white;
			border: none;
			padding: 10px 20px;
			border-radius: 6px;
			cursor: pointer;
			font-size: 14px;
			font-weight: 600;
			transition: background-color 0.2s;
		}
		.todo-btn:hover {
			background-color: #0056b3;
		}
		.char-counter {
			font-size: 12px;
			color: #888;
			margin-bottom: 16px;
			text-align: right;
		}
		ul {
			list-style-type: none;
			padding: 0;
			margin: 0;
		}
		li {
			background: #f8f9fa;
			padding: 12px 16px;
			border: 1px solid #e9ecef;
			border-radius: 6px;
			margin-bottom: 8px;
			font-size: 14px;
			display: flex;
			align-items: center;
			word-break: break-word;
		}
		li::before {
			content: "📌";
			margin-right: 10px;
			font-size: 14px;
		}
		.empty-msg {
			color: #888;
			font-style: italic;
			text-align: center;
			padding: 16px;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Hello from Kubernetes!</h1>
		<p>Hourly picture & Todo manager</p>
		<img src="/image.jpg" alt="Hourly Random Image" />
		
		<div class="todo-section">
			<h3>Todo List</h3>
			<form id="todoForm" class="todo-form">
				<input id="todoInput" class="todo-input" type="text" maxlength="140" placeholder="Enter a new todo (max 140 chars)..." required />
				<button class="todo-btn" type="submit">Send</button>
			</form>
			<div id="charCounter" class="char-counter">0 / 140</div>
			<ul id="todoList">
				<li class="empty-msg">Loading todos...</li>
			</ul>
		</div>
	</div>

	<script>
		const todoForm = document.getElementById('todoForm');
		const todoInput = document.getElementById('todoInput');
		const todoList = document.getElementById('todoList');
		const charCounter = document.getElementById('charCounter');

		todoInput.addEventListener('input', () => {
			charCounter.textContent = todoInput.value.length + ' / 140';
		});

		async function fetchTodos() {
			try {
				const res = await fetch('/todos');
				if (!res.ok) throw new Error('Network response was not ok');
				const data = await res.json();
				renderTodos(data);
			} catch (err) {
				console.error('Failed to load todos:', err);
				todoList.innerHTML = '<li class="empty-msg" style="color: #dc3545;">⚠️ Failed to load todos from backend.</li>';
			}
		}

		function renderTodos(todos) {
			if (!todos || todos.length === 0) {
				todoList.innerHTML = '<li class="empty-msg">No todos yet. Add one above!</li>';
				return;
			}
			todoList.innerHTML = '';
			todos.forEach(item => {
				const li = document.createElement('li');
				li.textContent = typeof item === 'object' && item.text ? item.text : item;
				todoList.appendChild(li);
			});
		}

		todoForm.addEventListener('submit', async (e) => {
			e.preventDefault();
			const text = todoInput.value.trim();
			if (!text) return;
			if (text.length > 140) {
				alert('Todo cannot exceed 140 characters.');
				return;
			}

			try {
				const res = await fetch('/todos', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ text: text })
				});

				if (!res.ok) {
					const errorText = await res.text();
					alert('Failed to save todo: ' + errorText);
					return;
				}

				todoInput.value = '';
				charCounter.textContent = '0 / 140';
				await fetchTodos();
			} catch (err) {
				console.error('Error posting todo:', err);
				alert('Error connecting to todo-backend.');
			}
		});

		// Initial fetch
		fetchTodos();
	</script>
</body>
</html>`
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Todo-app (frontend) started on port %s\n", port)

	http.HandleFunc("/", handler)
	http.HandleFunc("/image.jpg", imageHandler)
	http.HandleFunc("/crash", crashHandler)
	http.HandleFunc("/healthz", healthHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
