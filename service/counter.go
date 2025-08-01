package service

import (
	"database/sql"
	"sync"
	"time"

	"view-counter/database"
)

// 承载每日访问量统计结果
type DailyViewResult struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

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

// IncrementView 只负责业务逻辑，不再处理 HTTP
func (s *CounterService) IncrementView(domain string) error {
	currentDate := time.Now().UTC().Format("2006-01-02")

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // 如果后续出错，自动回滚

	// 更新日统计
	_, err = tx.Exec(`
		INSERT INTO daily_views (domain, date, count) 
		VALUES (?, ?, 1)
		ON CONFLICT(domain, date) DO UPDATE SET count = count + 1
	`, domain, currentDate)
	if err != nil {
		return err
	}

	// 更新总统计
	_, err = tx.Exec(`
		INSERT INTO total_views (domain, count)
		VALUES (?, 1)
		ON CONFLICT(domain) DO UPDATE SET count = count + 1
	`, domain)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// GetView 只负责业务逻辑，不再处理 HTTP
func (s *CounterService) GetView(domain string) (int, error) {
	var total int
	err := s.db.QueryRow(`
		SELECT count FROM total_views WHERE domain = ?
	`, domain).Scan(&total)

	// 将错误传递给上层(handler)处理
	if err != nil {
		return 0, err
	}

	return total, nil
}

// 获取指定域名在某个日期范围内的每日访问量
func (s *CounterService) GetDailyViews(domain, startDate, endDate string) ([]DailyViewResult, error) {
	rows, err := s.db.Query(`
		SELECT date, count FROM daily_views
		WHERE domain = ? AND date BETWEEN ? AND ?
		ORDER BY date ASC
	`, domain, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DailyViewResult
	for rows.Next() {
		var dv DailyViewResult
		if err := rows.Scan(&dv.Date, &dv.Count); err != nil {
			return nil, err
		}
		results = append(results, dv)
	}

	return results, nil
}