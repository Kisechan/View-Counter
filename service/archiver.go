package service

import (
	"database/sql"
	"log"
	"time"
)

type Archiver struct {
	db        *sql.DB
	archiveDB *sql.DB
	interval  time.Duration
	retention time.Duration
}

func NewArchiver(db *sql.DB, archiveDBPath string, interval, retention time.Duration) *Archiver {
	archiveDB, err := sql.Open("sqlite3", archiveDBPath)
	if err != nil {
		log.Fatalf("Failed to open archive database: %v", err)
	}

	// 初始化归档表
	_, err = archiveDB.Exec(`
		CREATE TABLE IF NOT EXISTS archived_daily_views (
			domain TEXT NOT NULL,
			date TEXT NOT NULL,
			count INTEGER DEFAULT 0,
			PRIMARY KEY (domain, date)
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create archive table: %v", err)
	}

	return &Archiver{
		db:        db,
		archiveDB: archiveDB,
		interval:  interval,
		retention: retention,
	}
}

func (a *Archiver) Start() {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for range ticker.C {
		a.archiveOldData()
	}
}

func (a *Archiver) archiveOldData() {
	cutoffDate := time.Now().UTC().Add(-a.retention).Format("2006-01-02")

	tx, err := a.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// 将旧数据复制到归档数据库
	_, err = tx.Exec(`
		INSERT INTO archived_daily_views
		SELECT * FROM daily_views WHERE date < ?
	`, cutoffDate)
	if err != nil {
		log.Printf("Failed to copy to archive: %v", err)
		return
	}

	// 删除原数据库中的旧数据
	_, err = tx.Exec(`
		DELETE FROM daily_views WHERE date < ?
	`, cutoffDate)
	if err != nil {
		log.Printf("Failed to delete old data: %v", err)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit archive transaction: %v", err)
		return
	}

	log.Printf("Successfully archived data older than %s", cutoffDate)
}
