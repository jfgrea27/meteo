package config

import (
	"github.com/jfgrea27/meteo/internal/utils"
)

type Config struct {
	AppEnv string

	DBProvider  string
	DatabaseURL string
	DBSchema    string

	Port string
}

func Load() Config {
	port := utils.GetEnvVar("API_PORT", false)
	if port == "" {
		port = "8080"
	}

	return Config{
		AppEnv: utils.GetEnvVar("APP_ENV", false),

		DBProvider:  utils.GetEnvVar("DB_PROVIDER", true),
		DatabaseURL: utils.GetEnvVar("DATABASE_URL", true),
		DBSchema:    utils.GetEnvVar("DB_SCHEMA", false),

		Port: port,
	}
}
