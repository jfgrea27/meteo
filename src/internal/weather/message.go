package weather

import "encoding/json"

type WeatherMessage struct {
	Provider Provider        `json:"provider"`
	City     City            `json:"city"`
	Content  json.RawMessage `json:"content"`
}
