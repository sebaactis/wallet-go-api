package config

import "os"

type Config struct {
	HTTPAddr string
	Driver   string
	DSN      string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}

func Load() Config {
	return Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Driver:   getEnv("DB_DRIVER", "sqlite"),
		DSN:      getEnv("DB_DSN", "file:wallet.db?cache=shared&mode=rwc"),
	}
}