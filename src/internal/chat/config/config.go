package config

import (
	"github.com/jfgrea27/rt-meteo/internal/utils"
)

type Config struct {
	AppEnv string

	DBProvider  string
	DatabaseURL string
	DBSchema    string

	Port string

	AIProvider string
	AIModel    string
	AIEndpoint string
}

func Load() Config {
	port := utils.GetEnvVar("CHAT_PORT", false)
	if port == "" {
		port = "50051"
	}

	return Config{
		AppEnv: utils.GetEnvVar("APP_ENV", false),

		DBProvider:  utils.GetEnvVar("DB_PROVIDER", true),
		DatabaseURL: utils.GetEnvVar("DATABASE_URL", true),
		DBSchema:    utils.GetEnvVar("DB_SCHEMA", false),

		Port: port,

		AIProvider: utils.GetEnvVar("AI_PROVIDER", true),
		AIModel:    utils.GetEnvVar("AI_MODEL", true),
		AIEndpoint: utils.GetEnvVar("AI_ENDPOINT", true),
	}
}
