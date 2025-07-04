package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
	mu sync.Mutex
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./views.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS counter (
		domain TEXT PRIMARY KEY,
		total INTEGER
	);`)
	if err != nil {
		log.Fatal(err)
	}
}

func extractDomain(r *http.Request) string {
	// 优先使用 nginx 传递的真实主机名
	host := r.Header.Get("X-Real-Host")
	if host == "" {
		// 备选方式：使用 Host 头部
		host = r.Host
	}
	return strings.ToLower(strings.Split(host, ":")[0])
}

func incrementView(w http.ResponseWriter, r *http.Request) {
	domain := extractDomain(r)

	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(`
		INSERT INTO counter (domain, total) VALUES (?, 1)
		ON CONFLICT(domain) DO UPDATE SET total = total + 1
	`, domain)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getView(w http.ResponseWriter, r *http.Request) {
	domain := extractDomain(r)

	var total int
	err := db.QueryRow(`SELECT total FROM counter WHERE domain = ?`, domain).Scan(&total)
	if err == sql.ErrNoRows {
		total = 0
	} else if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%d", total)
}

func main() {
	initDB()
	fmt.Println("Database initialized successfully")

	http.HandleFunc("/api/view", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			incrementView(w, r)
		case http.MethodGet:
			getView(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
