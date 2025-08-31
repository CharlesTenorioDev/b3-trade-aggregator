package config

import (
	"fmt"
	"os"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config/logger"
	"go.uber.org/zap"
)

const (
	DEVELOPER    = "developer"
	HOMOLOGATION = "homologation"
	PRODUCTION   = "production"
)

type Config struct {
	APPName     string `json:"app_name"`
	ServerHost  string `json:"server_host"`
	Port        string `json:"port"`
	APIPort     string `json:"api_port"`
	Mode        string `json:"mode"`
	DatabaseURL string `json:"database_url"`
	*PGSQLConfig
	FilePath string `json:"file_path"`
}

type PGSQLConfig struct {
	DB_DRIVE                  string `json:"db_drive"`
	DB_HOST                   string `json:"db_host"`
	DB_PORT                   string `json:"db_port"`
	DB_USER                   string `json:"db_user"`
	DB_PASS                   string `json:"db_pass"`
	DB_NAME                   string `json:"db_name"`
	DB_DSN                    string `json:"-"`
	DB_SET_MAX_OPEN_CONNS     int    `json:"db_set_max_open_conns"`
	DB_SET_MAX_IDLE_CONNS     int    `json:"db_set_max_idle_conns"`
	DB_SET_CONN_MAX_LIFE_TIME int    `json:"db_set_conn_max_life_time"`
	SRV_DB_SSL_MODE           string `json:"srv_db_ssl_mode"`
}

// LoadConfig carrega as configurações de variáveis de ambiente.
func LoadConfig() *Config {
	cfg := NewConfig()

	// Set API port from environment or use default
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
		logger.Info("API_PORT não definida, usando default", zap.String("default_port", apiPort))
	}
	cfg.APIPort = apiPort

	// Set database URL from environment or construct from components
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Read individual database components from environment
		dbHost := os.Getenv("SRV_DB_HOST")
		dbPort := os.Getenv("SRV_DB_PORT")
		dbUser := os.Getenv("SRV_DB_USER")
		dbPass := os.Getenv("SRV_DB_PASS")
		dbName := os.Getenv("SRV_DB_NAME")
		dbSSLMode := os.Getenv("SRV_DB_SSL_MODE")

		// Update the PGSQLConfig with the values from environment
		cfg.DB_HOST = dbHost
		cfg.DB_PORT = dbPort
		cfg.DB_USER = dbUser
		cfg.DB_PASS = dbPass
		cfg.DB_NAME = dbName
		cfg.SRV_DB_SSL_MODE = dbSSLMode

		// Construct the DATABASE_URL
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPass, dbHost, dbPort, dbName, dbSSLMode)
		logger.Info("DATABASE_URL não definida, usando construída", zap.String("constructed_url", dbURL))
	}
	cfg.DatabaseURL = dbURL

	return cfg
}

func NewConfig() *Config {
	conf := defaultConf()

	SRV_PORT := os.Getenv("SRV_PORT")
	if SRV_PORT != "" {
		conf.Port = SRV_PORT
	}

	SRV_MODE := os.Getenv("SRV_MODE")
	if SRV_MODE != "" {
		conf.Mode = SRV_MODE
	}

	SRV_DB_SSL_MODE := os.Getenv("SRV_DB_SSL_MODE")
	if SRV_DB_SSL_MODE != "" {
		conf.PGSQLConfig.SRV_DB_SSL_MODE = SRV_DB_SSL_MODE
	}

	return conf
}

func defaultConf() *Config {
	default_conf := Config{
		Port: "8080",
		Mode: DEVELOPER,
		// 7 days for refresh token (7 * 24 * 60 = 10080 minutes)

		PGSQLConfig: &PGSQLConfig{
			DB_DRIVE: "postgres",
			DB_PORT:  "5432",
		},
	}

	return &default_conf
}
