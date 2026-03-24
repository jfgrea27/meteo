package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jfgrea27/meteo/internal/weather"
)

type Database interface {
	SaveWeatherEntry(ctx context.Context, entry *weather.WeatherEntry) error
	GetCurrentWeather(ctx context.Context, city weather.City) (*weather.WeatherEntry, error)
	GetHistoricalWeather(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error)
	Close() error
}

func ConstructDatabase(log *slog.Logger, provider, connStr, schema string) Database {
	switch provider {
	case "postgres":
		return ConstructPgDatabase(log, connStr, schema)
	default:
		panic(fmt.Sprintf("%s is not a valid database provider", provider))
	}
}
