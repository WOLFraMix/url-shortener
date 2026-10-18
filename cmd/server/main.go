package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/wolframix/url-shortener/internal/config"
	"github.com/wolframix/url-shortener/internal/handlers"
	"github.com/wolframix/url-shortener/internal/migrations"
	"github.com/wolframix/url-shortener/internal/storage"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	cfg := config.NewConfig()

	err = migrations.Run(cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	health := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Println(err)
		}
	}

	st, err := storage.NewStorage(cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/shorten", handlers.ShortenHandler(st))
	mux.HandleFunc("/{code}", handlers.GetByShortCodeHandler(st))

	log.Println("server listening on", cfg.Port)
	err = http.ListenAndServe(cfg.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
