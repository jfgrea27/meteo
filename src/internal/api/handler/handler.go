package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	api "github.com/jfgrea27/meteo/internal/api"
	"github.com/jfgrea27/meteo/internal/db"
	"github.com/jfgrea27/meteo/internal/weather"
)

type Handler struct {
	log *slog.Logger
	db  db.Database
}

func New(log *slog.Logger, db db.Database) *Handler {
	return &Handler{
		log: log.With("component", "api-handler"),
		db:  db,
	}
}

func (h *Handler) GetCurrentWeather(c *gin.Context) {
	city := strings.ToLower(c.Param("city"))
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city is required"})
		return
	}

	if _, ok := weather.CITY_COORDINATES[weather.City(city)]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported city"})
		return
	}

	entry, err := h.db.GetCurrentWeather(c.Request.Context(), weather.City(city))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no weather data found"})
			return
		}
		h.log.Error("failed to get current weather", "city", city, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, api.ToWeatherResponse(entry))
}

func (h *Handler) GetHistoricalWeather(c *gin.Context) {
	city := strings.ToLower(c.Param("city"))
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city is required"})
		return
	}

	if _, ok := weather.CITY_COORDINATES[weather.City(city)]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported city"})
		return
	}

	var q api.HistoricalQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to query parameters are required (RFC3339 format)"})
		return
	}

	from, err := time.Parse(time.RFC3339, q.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' time format, use RFC3339"})
		return
	}

	to, err := time.Parse(time.RFC3339, q.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' time format, use RFC3339"})
		return
	}

	entries, err := h.db.GetHistoricalWeather(c.Request.Context(), weather.City(city), from, to)
	if err != nil {
		h.log.Error("failed to get historical weather", "city", city, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	result := make([]api.WeatherResponse, len(entries))
	for i, e := range entries {
		result[i] = api.ToWeatherResponse(&e)
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetCities(c *gin.Context) {
	cities := make([]string, 0, len(weather.CITY_COORDINATES))
	for city := range weather.CITY_COORDINATES {
		cities = append(cities, string(city))
	}
	c.JSON(http.StatusOK, gin.H{"cities": cities})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
