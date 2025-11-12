package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort        string
	DatabaseURL     string
	AdminToken      string
	UserToken       string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL не задан")
	}

	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		return Config{}, fmt.Errorf("ADMIN_TOKEN не задан")
	}

	userToken := os.Getenv("USER_TOKEN")
	if userToken == "" {
		return Config{}, fmt.Errorf("USER_TOKEN не задан")
	}

	logLevel := os.Getenv("LOG_LEVEL")

	shutdownTimeout := 10 * time.Second
	if v := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); v != "" {
		if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
			shutdownTimeout = time.Duration(seconds) * time.Second
		}
	}

	return Config{
		HTTPPort:        port,
		DatabaseURL:     databaseURL,
		AdminToken:      adminToken,
		UserToken:       userToken,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}
