package api

import (
	"time"

	"github.com/jfgrea27/meteo/internal/weather"
)

type WeatherResponse struct {
	Time        time.Time `json:"time"`
	City        string    `json:"city"`
	Temperature float32   `json:"temperature"`
	Pressure    float32   `json:"pressure"`
	Humidity    float32   `json:"humidity"`
	WindSpeed   float32   `json:"wind_speed"`
	UV          float32   `json:"uv"`
	Description string    `json:"description"`
}

func ToWeatherResponse(entry *weather.WeatherEntry) WeatherResponse {
	return WeatherResponse{
		Time:        entry.Time,
		City:        string(entry.City),
		Temperature: entry.Temperature,
		Pressure:    entry.Pressure,
		Humidity:    entry.Humidity,
		WindSpeed:   entry.WindSpeed,
		UV:          entry.UV,
		Description: entry.Description,
	}
}

type HistoricalQuery struct {
	From string `form:"from" binding:"required"`
	To   string `form:"to" binding:"required"`
}
