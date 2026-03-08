package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/jfgrea27/rt-meteo/internal/api/handler"
	"github.com/jfgrea27/rt-meteo/internal/db"
)

func New(log *slog.Logger, database db.Database) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	h := handler.New(log, database)

	r.GET("/health", h.HealthCheck)

	v1 := r.Group("/v1")
	{
		v1.GET("/cities", h.GetCities)
		v1.GET("/weather/:city/current", h.GetCurrentWeather)
		v1.GET("/weather/:city/historical", h.GetHistoricalWeather)
	}

	return r
}
