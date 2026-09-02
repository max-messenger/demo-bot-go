package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":19090", "stub server listen address")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/me", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"user_id":  1,
			"name":     "test_bot",
			"username": "test_bot",
		})
	})

	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, map[string]any{
				"subscriptions": []any{},
			})
		case http.MethodDelete:
			writeJSON(w, map[string]any{
				"success": true,
			})
		case http.MethodPost:
			writeJSON(w, map[string]any{
				"success": true,
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{})
	})

	log.Printf("stub server listening on %s", *addr)

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("stub server: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	if _, err := w.Write(data); err != nil {
		log.Printf("write response: %v", err)
	}
}
