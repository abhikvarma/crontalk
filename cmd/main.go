package main

import (
	"github.com/abhikvarma/crontalk/config"
	"github.com/abhikvarma/crontalk/internal/anthropic"
	"github.com/abhikvarma/crontalk/internal/api"
	"log"
	"net/http"
	"os"
)

func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.AnthropicApiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY env var not set")
	}
	if cfg.AnthropicModel == "" {
		log.Fatal("ANTHROPIC_MODEL env var not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	anthropicService := anthropic.NewService(cfg.AnthropicApiKey, cfg.AnthropicModel)
	handler := api.NewHandler(anthropicService)

	http.HandleFunc("/v1/cron", EnableCORS(handler.HandleCronRequest))

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
