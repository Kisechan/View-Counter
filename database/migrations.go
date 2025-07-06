package database

func runMigrations() error {
	_, err := db.Exec(`
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
		return err
	}
	return nil
}