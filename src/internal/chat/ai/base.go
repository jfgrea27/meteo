package ai

import (
	"fmt"
	"log/slog"
)

type AIChat interface {
	PreparePrompt(string) (string, error)
	Chat(string) (string, error)
}

func ConstructAiChat(log *slog.Logger, provider, endpoint, model string) AIChat {
	switch provider {
	case "ollama":
		return &Ollama{log: log, endpoint: endpoint, model: model}
	default:
		panic(fmt.Sprintf("%s is not a valid ai provider", provider))
	}
}
