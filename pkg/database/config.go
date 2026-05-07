package database

import (
	"os"
	"strconv"
	"time"
)

const defaultTimeout = 10 * time.Second

type Config struct {
	DatabaseURL string
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	Timeout     time.Duration
}

func LoadConfigFromEnv() Config {
	cfg := Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "eva_trainer",
		SSLMode:  "disable",
		Timeout:  defaultTimeout,
	}

	if url := os.Getenv("DATABASE_URL"); url != "" {
		cfg.DatabaseURL = url
		return cfg
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.SSLMode = v
	}
	if v := os.Getenv("DB_TIMEOUT"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			cfg.Timeout = time.Duration(sec) * time.Second
		}
	}

	return cfg
}
