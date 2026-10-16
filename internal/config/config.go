package config

import (
	"os"
	"strings"
)

type Config struct {
	Port string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // дефолт
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return &Config{
		Port: port,
	}
}
