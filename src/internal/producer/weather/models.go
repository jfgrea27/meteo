package weather

import "github.com/jfgrea27/meteo/internal/weather"

// Re-export shared types used by the producer config and callers.
type WeatherProvider = weather.Provider
type City = weather.City
type Coordinate = weather.Coordinate

const OpenWeather = weather.OpenWeather

var CITY_COORDINATES = weather.CITY_COORDINATES
