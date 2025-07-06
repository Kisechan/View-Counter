package config

type Config struct {
	DBPath string
}

func New() *Config {
	return &Config{
		DBPath: "./views.db",
	}
}