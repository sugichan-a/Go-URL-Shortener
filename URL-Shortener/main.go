package main

import (
    "html/template"
    "log"
    "math/rand"
    "net/http"
    "sync"
    "time"
)

var (
    urlStore = make(map[string]string)
    mu       sync.RWMutex
)

func generateShortID(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path

    if path == "/" {
        tmpl := template.Must(template.ParseFiles("static/index.html"))
        tmpl.Execute(w, nil)
    } else {
        shortID := path[1:]

        mu.RLock()
        originalURL, exists := urlStore[shortID]
        mu.RUnlock()

        if !exists {
            http.NotFound(w, r)
            return
        }

        http.Redirect(w, r, originalURL, http.StatusFound)
    }
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    url := r.FormValue("url")
    if url == "" {
        http.Error(w, "URL is required", http.StatusBadRequest)
        return
    }

    shortID := generateShortID(6)

    mu.Lock()
    urlStore[shortID] = url
    mu.Unlock()

    http.Redirect(w, r, "/?short="+shortID, http.StatusSeeOther)
}

func main() {
    rand.Seed(time.Now().UnixNano())

    http.HandleFunc("/shorten", shortenHandler)
    http.HandleFunc("/", rootHandler)

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
