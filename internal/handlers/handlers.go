package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Shortener interface {
	Save(originalURL, shortCode string) error
	GetByShortCode(code string) (string, error)
	GetByURL(originalURL string) (string, error)
}

type ShortenRequest struct {
	URL string `json:"url"`
}

func ParseShortenRequest(r *http.Request) (*ShortenRequest, error) {
	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func ValidateURL(raw string) error {
	if raw == "" {
		return errors.New("empty URL")
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return errors.New("err parse URL")
	}
	if u.Host == "" {
		return errors.New("missing host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid protocol")
	}
	return nil
}

func ShortenHandler(st Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := ParseShortenRequest(r)
		if err != nil {
			respondError(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if err := ValidateURL(req.URL); err != nil {
			respondError(w, "invalid url", http.StatusBadRequest)
			return
		}

		shortCode := GenerateShortCode(req.URL)

		err = st.Save(req.URL, shortCode)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				existing, getErr := st.GetByURL(req.URL)
				if getErr != nil {
					respondError(w, "internal server error", http.StatusInternalServerError)
					return
				}
				shortCode = existing
			} else {
				respondError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"short_code": shortCode})
	}
}

func GenerateShortCode(originalURL string) string {
	sum := sha256.Sum256([]byte(originalURL))
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return encoded[:8]
}

func GetByShortCodeHandler(st Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		originalUrl, err := st.GetByShortCode(code)
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			respondError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, originalUrl, http.StatusFound)
	}
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Println("write error response:", err)
	}
}
