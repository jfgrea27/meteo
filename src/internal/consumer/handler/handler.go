package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jfgrea27/meteo/internal/blob"
	"github.com/jfgrea27/meteo/internal/db"
	"github.com/jfgrea27/meteo/internal/weather"
)

type Handler struct {
	log  *slog.Logger
	blob blob.BlobStore
	db   db.Database
}

func New(log *slog.Logger, blob blob.BlobStore, db db.Database) *Handler {
	return &Handler{
		log:  log.With("component", "handler"),
		blob: blob,
		db:   db,
	}
}

func (h *Handler) Handle(body *string) error {
	var msg weather.WeatherMessage
	if err := json.Unmarshal([]byte(*body), &msg); err != nil {
		return fmt.Errorf("failed to unmarshal weather message: %w", err)
	}

	h.log.Info("received weather message", "provider", msg.Provider, "city", msg.City)

	// (i) Convert raw content to response objects
	resp, err := weather.ConvertCurrentWeather(msg.Provider, msg.Content)
	if err != nil {
		return fmt.Errorf("failed to convert weather data: %w", err)
	}

	// Use the city from the message envelope, not from the raw provider response
	resp.City = msg.City

	h.log.Info("converted weather data",
		"city", resp.City,
		"temp", resp.Temperature,
		"description", resp.Description,
	)

	// (ii) Save raw data to blob store
	key := fmt.Sprintf("%s/%s/%d.json", msg.Provider, msg.City, resp.Time.Unix())
	if err := h.blob.Save(key, msg.Content); err != nil {
		return fmt.Errorf("failed to save raw data to blob store: %w", err)
	}

	// (iii) Save converted objects to database
	if err := h.db.SaveWeatherEntry(context.TODO(), resp); err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	return nil
}
