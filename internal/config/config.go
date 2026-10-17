package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	Port  string
	DbUrl string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // дефолт
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("missing DATABASE_URL environment variable")
	}

	return &Config{
		Port:  port,
		DbUrl: dbUrl,
	}
}
