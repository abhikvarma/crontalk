package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	AnthropicApiKey string
	AnthropicModel  string
	port            string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return loadFromEnv(), nil
}

func loadFromEnv() *Config {
	return &Config{
		os.Getenv("ANTHROPIC_API_KEY"),
		os.Getenv("ANTHROPIC_MODEL"),
		getEnvOrDefault("PORT", "8080"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
