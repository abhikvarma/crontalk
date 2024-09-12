package main

import (
	"github.com/abhikvarma/crontalk/config"
	"github.com/abhikvarma/crontalk/internal/anthropic"
	"github.com/abhikvarma/crontalk/internal/api"
	"log"
	"net/http"
	"os"
)

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

	http.HandleFunc("/v1/cron", handler.HandleCronRequest)

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
