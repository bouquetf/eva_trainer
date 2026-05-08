package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"eva_trainer/pkg/database"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port    string
	GinMode string
}

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
}

type JWTConfig struct {
	Secret             string
	Expiry             time.Duration
	RefreshTokenExpiry time.Duration
}

type CORSConfig struct {
	Origins []string
}

type AppConfig struct {
	Server   ServerConfig
	Database database.Config
	Discord  DiscordConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

func Load() (*AppConfig, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg := &AppConfig{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Discord: DiscordConfig{
			ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		},
		JWT: JWTConfig{
			Secret:             os.Getenv("JWT_SECRET"),
			Expiry:             getDuration("JWT_EXPIRY", 24*time.Hour),
			RefreshTokenExpiry: getDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		Database: database.LoadConfigFromEnv(),
		CORS: CORSConfig{
			Origins: getOrigins("CORS_ORIGINS"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *AppConfig) Validate() error {
	var missing []string

	if c.JWT.Secret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.Discord.ClientID == "" {
		missing = append(missing, "DISCORD_CLIENT_ID")
	}
	if c.Discord.ClientSecret == "" {
		missing = append(missing, "DISCORD_CLIENT_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func getOrigins(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return []string{"*"}
	}

	parts := strings.Split(v, ",")
	origins := make([]string, len(parts))
	for i, p := range parts {
		origins[i] = strings.TrimSpace(p)
	}
	return origins
}

func NewDatabaseConfigFromApp(cfg *AppConfig) database.Config {
	return cfg.Database
}
