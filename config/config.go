package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	LogLevel        string
	Message         string
	HTTPPort        string
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		LogLevel:        os.Getenv("LOG_LEVEL"),
		Message:         fallback(os.Getenv("APP_MESSAGE"), "Привет!"),
		HTTPPort:        fallback(os.Getenv("HTTP_PORT"), "8080"),
		ShutdownTimeout: parseShutdownTimeout(os.Getenv("SHUTDOWN_TIMEOUT_SECONDS")),
	}
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

func parseShutdownTimeout(value string) time.Duration {
	if value == "" {
		return 5 * time.Second
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
