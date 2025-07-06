package service

import (
	"database/sql"
	"net/http"
	"sync"
	"time"
	"fmt"

	"view-counter/database"
	"view-counter/utils"
)

type CounterService struct {
	db *sql.DB
	mu *sync.Mutex
}

func NewCounterService(db *sql.DB) *CounterService {
	return &CounterService{
		db: db,
		mu: database.GetMutex(),
	}
}

func (s *CounterService) IncrementView(w http.ResponseWriter, r *http.Request) {
	domain := utils.ExtractDomain(r)
	if domain == "" {
		http.Error(w, "Domain not found", http.StatusBadRequest)
		return
	}

	currentDate := time.Now().UTC().Format("2006-01-02")
	
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
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

	if err = tx.Commit(); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (s *CounterService) GetView(w http.ResponseWriter, r *http.Request) {
	domain := utils.ExtractDomain(r)
	if domain == "" {
		http.Error(w, "Domain not found", http.StatusBadRequest)
		return
	}

	var total int
	err := s.db.QueryRow(`
		SELECT count FROM total_views WHERE domain = ?
	`, domain).Scan(&total)
	if err == sql.ErrNoRows {
		total = 0
	} else if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%d", total)))
}