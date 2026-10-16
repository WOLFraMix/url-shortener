package main

import (
	"log"
	"net/http"
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

	log.Println("server listening on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
