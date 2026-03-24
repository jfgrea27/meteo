package main

import (
	"context"
	"os"
	"time"

	"github.com/jfgrea27/meteo/internal/logger"
	"github.com/jfgrea27/meteo/internal/producer/config"
	"github.com/jfgrea27/meteo/internal/producer/weather"
	"github.com/jfgrea27/meteo/internal/queue"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("starting producer")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// setup weather service
	log.Info("initialising weather service", "provider", string(cfg.WeatherProvider))
	weatherSvc := weather.ConstructWeatherService(log, cfg.WeatherProvider, cfg.OpenWeatherAPIKey)

	// setup queue service
	log.Info("initialising queue service", "provider", string(cfg.QueueProvider), "queue", cfg.QueueName)
	queueSvc := queue.ConstructQueueService(cfg.QueueProvider, cfg.AWSAccount, cfg.AWSRegion, 0)

	messages := weather.AggregateCurrentWeather(log, weatherSvc)
	if len(messages) == 0 {
		log.Error("no weather data collected, skipping publish")
		os.Exit(1)
	}

	log.Info("publishing weather data", "count", len(messages), "queue", cfg.QueueName)
	for _, msg := range messages {
		if err := queueSvc.Produce(ctx, msg, cfg.QueueName); err != nil {
			log.Error("failed to publish weather data", "error", err)
			os.Exit(1)
		}
	}

	log.Info("producer finished successfully")
}
