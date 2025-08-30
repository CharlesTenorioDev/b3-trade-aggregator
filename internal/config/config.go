// internal/config/config.go (Exemplo simples)
package config

import (
	"log"
	"os"
)

// Config contém as configurações da aplicação.
type Config struct {
	DatabaseURL string
	APIPort     string
}

// LoadConfig carrega as configurações de variáveis de ambiente.
func LoadConfig() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/mydatabase?sslmode=disable" // Default para desenvolvimento
		log.Printf("DATABASE_URL não definida, usando default: %s", dbURL)
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
		log.Printf("API_PORT não definida, usando default: %s", apiPort)
	}

	return &Config{
		DatabaseURL: dbURL,
		APIPort:     apiPort,
	}
}
