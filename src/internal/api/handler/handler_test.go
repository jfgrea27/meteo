package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	api "github.com/jfgrea27/rt-meteo/internal/api"
	"github.com/jfgrea27/rt-meteo/internal/weather"
)

type mockDatabase struct {
	getCurrentFunc    func(ctx context.Context, city weather.City) (*weather.WeatherEntry, error)
	getHistoricalFunc func(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error)
}

func (m *mockDatabase) SaveWeatherEntry(ctx context.Context, entry *weather.WeatherEntry) error {
	return nil
}

func (m *mockDatabase) GetCurrentWeather(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
	if m.getCurrentFunc != nil {
		return m.getCurrentFunc(ctx, city)
	}
	return nil, nil
}

func (m *mockDatabase) GetHistoricalWeather(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
	if m.getHistoricalFunc != nil {
		return m.getHistoricalFunc(ctx, city, from, to)
	}
	return nil, nil
}

func (m *mockDatabase) Close() error {
	return nil
}

func setupRouter(db *mockDatabase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(slog.Default(), db)

	r.GET("/health", h.HealthCheck)
	r.GET("/v1/cities", h.GetCities)
	r.GET("/v1/weather/:city/current", h.GetCurrentWeather)
	r.GET("/v1/weather/:city/historical", h.GetHistoricalWeather)

	return r
}

func TestHealthCheck(t *testing.T) {
	r := setupRouter(&mockDatabase{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestGetCities(t *testing.T) {
	r := setupRouter(&mockDatabase{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/cities", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string][]string
	json.Unmarshal(w.Body.Bytes(), &body)
	cities := body["cities"]
	if len(cities) != len(weather.CITY_COORDINATES) {
		t.Errorf("expected %d cities, got %d", len(weather.CITY_COORDINATES), len(cities))
	}
}

func TestGetCurrentWeather(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		db := &mockDatabase{
			getCurrentFunc: func(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
				return &weather.WeatherEntry{
					Time:        fixedTime,
					City:        city,
					Temperature: 18.0,
					Pressure:    1015.0,
					Humidity:    65.0,
					WindSpeed:   6.0,
					UV:          3.0,
					Description: "few clouds",
				}, nil
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/current", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body api.WeatherResponse
		json.Unmarshal(w.Body.Bytes(), &body)
		if body.City != "london" {
			t.Errorf("city = %q, want london", body.City)
		}
		if body.Temperature != 18.0 {
			t.Errorf("temperature = %v, want 18.0", body.Temperature)
		}
	})

	t.Run("uppercase city is lowercased", func(t *testing.T) {
		db := &mockDatabase{
			getCurrentFunc: func(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
				if city != "london" {
					t.Errorf("expected city to be lowercased to 'london', got %q", city)
				}
				return &weather.WeatherEntry{
					Time:        fixedTime,
					City:        city,
					Temperature: 18.0,
				}, nil
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/London/current", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unsupported city", func(t *testing.T) {
		r := setupRouter(&mockDatabase{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/UnknownCity/current", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("no data found", func(t *testing.T) {
		db := &mockDatabase{
			getCurrentFunc: func(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
				return nil, fmt.Errorf("failed to get current weather for %s: %w", city, sql.ErrNoRows)
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/current", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := &mockDatabase{
			getCurrentFunc: func(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
				return nil, errors.New("connection refused")
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/current", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestGetHistoricalWeather(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		db := &mockDatabase{
			getHistoricalFunc: func(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
				return []weather.WeatherEntry{
					{Time: fixedTime, City: city, Temperature: 18.0, Description: "clear"},
					{Time: fixedTime.Add(time.Hour), City: city, Temperature: 19.0, Description: "cloudy"},
				}, nil
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body []api.WeatherResponse
		json.Unmarshal(w.Body.Bytes(), &body)
		if len(body) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(body))
		}
	})

	t.Run("missing query params", func(t *testing.T) {
		r := setupRouter(&mockDatabase{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid from format", func(t *testing.T) {
		r := setupRouter(&mockDatabase{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical?from=bad&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid to format", func(t *testing.T) {
		r := setupRouter(&mockDatabase{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical?from=2024-01-01T00:00:00Z&to=bad", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("uppercase city is lowercased", func(t *testing.T) {
		db := &mockDatabase{
			getHistoricalFunc: func(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
				if city != "london" {
					t.Errorf("expected city to be lowercased to 'london', got %q", city)
				}
				return []weather.WeatherEntry{
					{Time: fixedTime, City: city, Temperature: 18.0},
				}, nil
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/London/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unsupported city", func(t *testing.T) {
		r := setupRouter(&mockDatabase{})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/UnknownCity/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := &mockDatabase{
			getHistoricalFunc: func(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
				return nil, errors.New("connection refused")
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		db := &mockDatabase{
			getHistoricalFunc: func(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
				return []weather.WeatherEntry{}, nil
			},
		}

		r := setupRouter(db)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/weather/london/historical?from=2024-01-01T00:00:00Z&to=2024-01-02T00:00:00Z", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body []api.WeatherResponse
		json.Unmarshal(w.Body.Bytes(), &body)
		if len(body) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(body))
		}
	})
}
