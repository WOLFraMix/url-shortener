package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
	"github.com/wolframix/url-shortener/internal/config"
	"github.com/wolframix/url-shortener/internal/handlers"
	"github.com/wolframix/url-shortener/internal/storage"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	cfg := config.NewConfig()

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
	mux.HandleFunc("/shorten", shortenHandler(st))
	mux.HandleFunc("/{code}", getByShortCodeHandler(st))

	log.Println("server listening on", cfg.Port)
	err = http.ListenAndServe(cfg.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func shortenHandler(st *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := handlers.ParseShortenRequest(r)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		shortCode := generateShortCode(req.URL)

		err = st.Save(req.URL, shortCode)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				existing, getErr := st.GetByURL(req.URL)
				if getErr != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				shortCode = existing
			} else {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"short_code": shortCode})
	}
}

func generateShortCode(originalURL string) string {
	sum := sha256.Sum256([]byte(originalURL))
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return encoded[:8]
}

func getByShortCodeHandler(st *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		url, err := st.GetByShortCode(code)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}
