package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	BackendUrl string
	AgentID    string
	AgentToken string
}

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		BackendUrl: getEnv("BACKEND_URL", "http://localhost:8080"),
		AgentID:    getEnv("AGENT_ID", ""),
		AgentToken: getEnv("AGENT_TOKEN", ""),

	}
	if cfg.AgentID == "" || cfg.AgentToken == "" {
		log.Fatal("AGENT_ID and AGENT_TOKEN must be set")
	}
	return cfg

}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}