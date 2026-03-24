package weather

import (
	"log/slog"

	sharedweather "github.com/jfgrea27/meteo/internal/weather"
)

func AggregateCurrentWeather(log *slog.Logger, svc WeatherService) []sharedweather.WeatherMessage {
	messages := make([]sharedweather.WeatherMessage, 0, len(CITY_COORDINATES))

	for city := range CITY_COORDINATES {
		raw, err := svc.GetCurrentWeatherRaw(city)
		if err != nil {
			log.Error("failed to get weather, skipping city", "city", city, "error", err)
			continue
		}
		messages = append(messages, sharedweather.WeatherMessage{
			Provider: svc.Provider(),
			City:     city,
			Content:  raw,
		})
	}

	log.Info("aggregated weather data", "cities_ok", len(messages), "cities_total", len(CITY_COORDINATES))
	return messages
}
