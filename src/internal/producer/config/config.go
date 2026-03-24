package config

import (
	"github.com/jfgrea27/meteo/internal/producer/weather"
	"github.com/jfgrea27/meteo/internal/queue"
	"github.com/jfgrea27/meteo/internal/utils"
)

type Config struct {
	AppEnv string

	// Required — the producer cannot run without knowing which providers to use.
	WeatherProvider weather.WeatherProvider
	QueueProvider   queue.QueueProvider
	QueueName       string

	// Provider details — not required at startup because only the active
	// provider needs its credentials. Validated when the provider is constructed.
	OpenWeatherAPIKey string
	AWSAccount        string
	AWSRegion         string
}

func Load() Config {
	return Config{
		AppEnv: utils.GetEnvVar("APP_ENV", false),

		WeatherProvider: weather.WeatherProvider(utils.GetEnvVar("WEATHER_PROVIDER", true)),
		QueueProvider:   queue.QueueProvider(utils.GetEnvVar("QUEUE_PROVIDER", true)),
		QueueName:       utils.GetEnvVar("QUEUE_NAME", true),

		OpenWeatherAPIKey: utils.GetEnvVar("OPENWEATHER_API_KEY", false),
		AWSAccount:        utils.GetEnvVar("AWS_ACCOUNT", false),
		AWSRegion:         utils.GetEnvVar("AWS_REGION", false),
	}
}
