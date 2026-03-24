package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	sharedweather "github.com/jfgrea27/meteo/internal/weather"
)

var (
	GET_CURRENT_WEATHER_API    = "https://api.openweathermap.org/data/2.5/weather?appid=%s&lat=%f&lon=%f&units=metric"
	GET_HISTORICAL_WEATHER_API = "https://history.openweathermap.org/data/2.5/history/city?appid=%s&lat=%f&lon=%f&type=hour&start=%d&end=%d"
)

type OpenWeatherService struct {
	apiKey string
	client *http.Client
	log    *slog.Logger
}

func (s *OpenWeatherService) Provider() sharedweather.Provider {
	return sharedweather.OpenWeather
}

func (s *OpenWeatherService) GetCurrentWeatherRaw(c City) (json.RawMessage, error) {
	coord := CITY_COORDINATES[c]

	s.log.Debug("fetching current weather", "city", c, "lat", coord.Lat, "lon", coord.Lon)

	api := fmt.Sprintf(GET_CURRENT_WEATHER_API, s.apiKey, coord.Lat, coord.Lon)

	resp, err := s.client.Get(api)
	if err != nil {
		s.log.Error("request failed", "city", c, "error", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("unexpected status code", "city", c, "status", resp.StatusCode)
		return nil, fmt.Errorf("status code %d != %d", resp.StatusCode, http.StatusOK)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("failed to read response body", "city", c, "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	s.log.Debug("fetched current weather raw", "city", c, "bytes", len(raw))
	return json.RawMessage(raw), nil
}

func (s *OpenWeatherService) GetHistoricalWeatherRaw(c City, from time.Time, to time.Time) (json.RawMessage, error) {
	coord := CITY_COORDINATES[c]

	s.log.Debug("fetching historical weather", "city", c, "from", from, "to", to)

	api := fmt.Sprintf(GET_HISTORICAL_WEATHER_API, s.apiKey, coord.Lat, coord.Lon, from.Unix(), to.Unix())

	resp, err := s.client.Get(api)
	if err != nil {
		s.log.Error("request failed", "city", c, "error", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("unexpected status code", "city", c, "status", resp.StatusCode)
		return nil, fmt.Errorf("status code %d != %d", resp.StatusCode, http.StatusOK)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("failed to read response body", "city", c, "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	s.log.Debug("fetched historical weather raw", "city", c, "bytes", len(raw))
	return json.RawMessage(raw), nil
}
