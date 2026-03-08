package main

import (
	"os"

	"github.com/jfgrea27/rt-meteo/internal/api/config"
	"github.com/jfgrea27/rt-meteo/internal/api/router"
	"github.com/jfgrea27/rt-meteo/internal/db"
	"github.com/jfgrea27/rt-meteo/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("starting api server", "port", cfg.Port)

	database := db.ConstructDatabase(log, cfg.DBProvider, cfg.DatabaseURL, cfg.DBSchema)
	defer database.Close()

	r := router.New(log, database)

	log.Info("listening", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
