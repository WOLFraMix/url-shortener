package handlers

import (
	"encoding/json"
	"net/http"
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
