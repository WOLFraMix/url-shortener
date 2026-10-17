package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

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
