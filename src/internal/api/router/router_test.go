package router

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jfgrea27/rt-meteo/internal/weather"
)

type mockDatabase struct{}

func (m *mockDatabase) SaveWeatherEntry(ctx context.Context, entry *weather.WeatherEntry) error {
	return nil
}

func (m *mockDatabase) GetCurrentWeather(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
	return &weather.WeatherEntry{
		Time:        time.Now(),
		City:        city,
		Temperature: 20.0,
		Description: "clear sky",
	}, nil
}

func (m *mockDatabase) GetHistoricalWeather(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
	return []weather.WeatherEntry{}, nil
}

func (m *mockDatabase) Close() error {
	return nil
}

func TestRoutes(t *testing.T) {
	r := New(slog.Default(), &mockDatabase{})

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"health", "GET", "/health", http.StatusOK},
		{"cities", "GET", "/v1/cities", http.StatusOK},
		{"current weather", "GET", "/v1/weather/London/current", http.StatusOK},
		{"historical weather", "GET", "/v1/weather/London/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", http.StatusOK},
		{"unknown route", "GET", "/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.path, w.Code, tt.status)
			}
		})
	}
}
