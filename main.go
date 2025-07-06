package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	
	// 创建表结构
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS daily_views (
			domain TEXT NOT NULL,
			date TEXT NOT NULL,
			count INTEGER DEFAULT 0,
			PRIMARY KEY (domain, date)
		);

		CREATE TABLE IF NOT EXISTS total_views (
			domain TEXT PRIMARY KEY,
			count INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_daily_domain ON daily_views(domain);
		CREATE INDEX IF NOT EXISTS idx_daily_date ON daily_views(date);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func extractDomain(r *http.Request) string {
	// 优先从 Referer 获取域名
	referer := r.Header.Get("Referer")
	if referer != "" {
		if strings.HasPrefix(referer, "http://") || strings.HasPrefix(referer, "https://") {
			host := strings.TrimPrefix(strings.TrimPrefix(referer, "http://"), "https://")
			host = strings.Split(host, "/")[0]
			host = strings.Split(host, ":")[0]
			return strings.ToLower(host)
		}
	}
	
	// 使用 X-Real-Host 和 Host头部检测域名
	host := r.Header.Get("X-Real-Host")
	if host == "" {
		host = r.Host
	}
	return strings.ToLower(strings.Split(host, ":")[0])
}

func incrementView(w http.ResponseWriter, r *http.Request) {
    domain := extractDomain(r)
    if domain == "" {
        http.Error(w, "Domain not found", http.StatusBadRequest)
        return
    }

    currentDate := time.Now().UTC().Format("2006-01-02")
    
    mu.Lock()
    defer mu.Unlock()

	// 开始事务
    tx, err := db.Begin()
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    // 更新日统计
    _, err = tx.Exec(`
        INSERT INTO daily_views (domain, date, count) 
        VALUES (?, ?, 1)
        ON CONFLICT(domain, date) DO UPDATE SET count = count + 1
    `, domain, currentDate)
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }

    // 更新总统计
    _, err = tx.Exec(`
        INSERT INTO total_views (domain, count)
        VALUES (?, 1)
        ON CONFLICT(domain) DO UPDATE SET count = count + 1
    `, domain)
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }

    err = tx.Commit()
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

func getView(w http.ResponseWriter, r *http.Request) {
    domain := extractDomain(r)
    if domain == "" {
        http.Error(w, "Domain not found", http.StatusBadRequest)
        return
    }

    var total int
    // 直接从 total_views 表获取，无需计算
    err := db.QueryRow(`
        SELECT count FROM total_views WHERE domain = ?
    `, domain).Scan(&total)
    if err == sql.ErrNoRows {
        total = 0 // 新域名，还没有记录
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