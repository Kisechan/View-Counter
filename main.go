package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS counter (id INTEGER PRIMARY KEY CHECK (id = 1), total INTEGER);`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`INSERT OR IGNORE INTO counter (id, total) VALUES (1, 0);`)
	if err != nil {
		log.Fatal(err)
	}
}

func incrementView(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(`UPDATE counter SET total = total + 1 WHERE id = 1;`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getView(w http.ResponseWriter, r *http.Request) {
	var total int
	err := db.QueryRow(`SELECT total FROM counter WHERE id = 1;`).Scan(&total)
	if err != nil {
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
		if r.Method == http.MethodPost {
			incrementView(w, r)
		} else if r.Method == http.MethodGet {
			getView(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
