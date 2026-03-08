package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jfgrea27/rt-meteo/internal/weather"
)

type mockBlobStore struct {
	saveFunc func(key string, data []byte) error
	lastKey  string
	lastData []byte
}

func (m *mockBlobStore) Save(key string, data []byte) error {
	m.lastKey = key
	m.lastData = data
	if m.saveFunc != nil {
		return m.saveFunc(key, data)
	}
	return nil
}

type mockDatabase struct {
	saveFunc  func(ctx context.Context, entry *weather.CurrentWeatherResponse) error
	lastEntry *weather.CurrentWeatherResponse
}

func (m *mockDatabase) SaveWeatherEntry(ctx context.Context, entry *weather.CurrentWeatherResponse) error {
	m.lastEntry = entry
	if m.saveFunc != nil {
		return m.saveFunc(ctx, entry)
	}
	return nil
}

func (m *mockDatabase) GetCurrentWeather(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
	return nil, nil
}

func (m *mockDatabase) GetHistoricalWeather(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
	return nil, nil
}

func (m *mockDatabase) Close() error {
	return nil
}

func validMessage() string {
	rawContent, _ := json.Marshal(map[string]interface{}{
		"dt": float64(1700000000),
		"main": map[string]interface{}{
			"temp":     float64(18.0),
			"pressure": float64(1015.0),
			"humidity": float64(65.0),
		},
		"wind": map[string]interface{}{
			"speed": float64(6.0),
		},
		"weather": []interface{}{
			map[string]interface{}{
				"description": "few clouds",
			},
		},
		"name": "London",
	})

	msg := weather.WeatherMessage{
		Provider: weather.OpenWeather,
		City:     "london",
		Content:  json.RawMessage(rawContent),
	}
	body, _ := json.Marshal(msg)
	return string(body)
}

func TestHandle(t *testing.T) {
	t.Run("valid openweather message", func(t *testing.T) {
		blob := &mockBlobStore{}
		db := &mockDatabase{}
		h := New(slog.Default(), blob, db)

		body := validMessage()
		err := h.Handle(&body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if blob.lastKey == "" {
			t.Error("expected blob Save to be called")
		}
		if db.lastEntry == nil {
			t.Fatal("expected db SaveWeatherEntry to be called")
		}
		if db.lastEntry.Temperature != 18.0 {
			t.Errorf("Temperature = %v, want 18.0", db.lastEntry.Temperature)
		}
		if db.lastEntry.City != "london" {
			t.Errorf("City = %q, want %q", db.lastEntry.City, "london")
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		h := New(slog.Default(), &mockBlobStore{}, &mockDatabase{})
		body := "not json"
		err := h.Handle(&body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("unsupported provider", func(t *testing.T) {
		h := New(slog.Default(), &mockBlobStore{}, &mockDatabase{})

		msg := weather.WeatherMessage{
			Provider: "unknown",
			Content:  json.RawMessage(`{"temp": 20.0}`),
		}
		body, _ := json.Marshal(msg)
		bodyStr := string(body)

		err := h.Handle(&bodyStr)
		if err == nil {
			t.Fatal("expected error for unsupported provider")
		}
	})

	t.Run("invalid raw content for provider", func(t *testing.T) {
		h := New(slog.Default(), &mockBlobStore{}, &mockDatabase{})

		msg := weather.WeatherMessage{
			Provider: weather.OpenWeather,
			Content:  json.RawMessage(`{"no_dt": true}`),
		}
		body, _ := json.Marshal(msg)
		bodyStr := string(body)

		err := h.Handle(&bodyStr)
		if err == nil {
			t.Fatal("expected error for invalid content")
		}
	})

	t.Run("blob store error", func(t *testing.T) {
		blobErr := errors.New("s3 unavailable")
		blob := &mockBlobStore{
			saveFunc: func(key string, data []byte) error { return blobErr },
		}
		h := New(slog.Default(), blob, &mockDatabase{})

		body := validMessage()
		err := h.Handle(&body)
		if !errors.Is(err, blobErr) {
			t.Fatalf("expected blob error, got %v", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("db unavailable")
		db := &mockDatabase{
			saveFunc: func(ctx context.Context, entry *weather.CurrentWeatherResponse) error { return dbErr },
		}
		h := New(slog.Default(), &mockBlobStore{}, db)

		body := validMessage()
		err := h.Handle(&body)
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected db error, got %v", err)
		}
	})
}
