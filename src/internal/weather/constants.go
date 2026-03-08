package weather

type Provider string

const (
	OpenWeather Provider = "openweather"
)

var CITY_COORDINATES = map[City]Coordinate{
	"london": {
		Lat: 51.5073219,
		Lon: -0.1276474,
	},
	"paris": {
		Lat: 48.8588897,
		Lon: 2.3200410217200766,
	},
}
