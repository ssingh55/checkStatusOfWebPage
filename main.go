package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"strings"
)

// Response struct defines the JSON response format
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/check", handleCheck)
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>URL Status Checker</title>
		<script>
			function checkUrl() {
				const urlInput = document.getElementById('urlInput').value;
				
				fetch('/check', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/x-www-form-urlencoded',
					},
					body: 'url=' + encodeURIComponent(urlInput)
				})
				.then(response => response.json())
				.then(data => {
					const resultDiv = document.getElementById('result');
					resultDiv.innerHTML = data.status === 200 ? 
						'✅ Status 200 OK' : 
						`+"`❌ Status ${data.status} (${data.message})`;"+`
					
					const historyItem = {
						url: urlInput,
						status: data.status,
						timestamp: new Date().toISOString()
					};
					
					const history = JSON.parse(sessionStorage.getItem('urlHistory') || '[]');
					history.push(historyItem);
					sessionStorage.setItem('urlHistory', JSON.stringify(history));
					
					updateHistoryDisplay();
				});
			}

			function updateHistoryDisplay() {
				const history = JSON.parse(sessionStorage.getItem('urlHistory') || '[]');
				const historyList = document.getElementById('historyList');
				historyList.innerHTML = history.reverse().map(item => 
					`+"`<li>${new Date(item.timestamp).toLocaleString()} - ${item.url} - Status: ${item.status}</li>`"+`
				).join('');
			}

			window.onload = updateHistoryDisplay;
		</script>
	</head>
	<body>
		<h1>URL Status Checker</h1>
		<input type="text" id="urlInput" placeholder="Enter URL" style="width: 300px;">
		<button onclick="checkUrl()">Check Status</button>
		<div id="result" style="margin: 20px 0; font-weight: bold;"></div>
		
		<h3>Check History (Session Storage)</h3>
		<ul id="historyList"></ul>
	</body>
	</html>
	`)
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawURL := r.FormValue("url")
	if rawURL == "" {
		sendJSONResponse(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Prevent checking the server itself
	if strings.Contains(rawURL, "localhost:8080") || strings.Contains(rawURL, "127.0.0.1:8080") {
		sendJSONResponse(w, http.StatusBadRequest, "Cannot check this server's status")
		return
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "Invalid URL format")
		return
	}

	// Configure proper HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}

	// Create proper request with headers
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	sendJSONResponse(w, resp.StatusCode, http.StatusText(resp.StatusCode))
}

func sendJSONResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status:  status,
		Message: message,
	})
}