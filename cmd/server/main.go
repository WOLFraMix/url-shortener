package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/wolframix/url-shortener/internal/config"
	"github.com/wolframix/url-shortener/internal/storage"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	cfg := config.NewConfig()

	_, err = storage.NewStorage(cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("db connected:")

	health := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Println(err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)

	log.Println("server listening on", cfg.Port)
	err = http.ListenAndServe(cfg.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
