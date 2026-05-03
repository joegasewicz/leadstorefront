package pkgs

import "os"

var Config = Load()

type AppConfig struct {
	Postgres Postgres
	Web      Service
	API      Service
	Session  Session
}

type Postgres struct {
	Host     string
	User     string
	Password string
	Database string
	Port     string
	SSLMode  string
}

type Service struct {
	Domain string
	Addr   string
}

type Session struct {
	Secret string
}

func Load() AppConfig {
	return AppConfig{
		Postgres: Postgres{
			Host:     envOrDefault("POSTGRES_HOST", "localhost"),
			User:     envOrDefault("POSTGRES_USER", "admin"),
			Password: envOrDefault("POSTGRES_PASSWORD", "admin"),
			Database: envOrDefault("POSTGRES_DB", "gadgetscout"),
			Port:     envOrDefault("POSTGRES_PORT", "5432"),
			SSLMode:  envOrDefault("POSTGRES_SSLMODE", "disable"),
		},
		Web: Service{
			Domain: envOrDefault("WEB_DOMAIN", "localhost"),
			Addr:   envOrDefault("WEB_ADDR", ":8000"),
		},
		API: Service{
			Domain: envOrDefault("API_DOMAIN", "localhost"),
			Addr:   envOrDefault("API_ADDR", ":8001"),
		},
		Session: Session{
			Secret: envOrDefault("SESSION_SECRET", "local-dev-session-secret-change-me"),
		},
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
