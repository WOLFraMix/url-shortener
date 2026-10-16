package main

import (
	"log"
	"net/http"

	"github.com/wolframix/url-shortener/internal/config"
)

func main() {
	health := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Println(err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)

	cfg := config.NewConfig()
	log.Println("server listening on", cfg.Port)
	err := http.ListenAndServe(cfg.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
