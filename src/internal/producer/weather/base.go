package weather

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sharedweather "github.com/jfgrea27/meteo/internal/weather"
)

type WeatherService interface {
	GetCurrentWeatherRaw(c City) (json.RawMessage, error)
	GetHistoricalWeatherRaw(c City, from time.Time, to time.Time) (json.RawMessage, error)
	Provider() sharedweather.Provider
}

func ConstructWeatherService(log *slog.Logger, p WeatherProvider, apiKey string) WeatherService {
	var svc WeatherService
	switch p {
	case OpenWeather:
		if apiKey == "" {
			panic("OPENWEATHER_API_KEY is required for openweather provider")
		}
		svc = &OpenWeatherService{
			apiKey: apiKey,
			client: &http.Client{},
			log:    log.With("service", "openweather"),
		}
	default:
		panic(fmt.Sprintf("%s is not a valid weather provider", p))
	}
	return svc
}
