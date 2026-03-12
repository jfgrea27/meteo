package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type Ollama struct {
	model    string
	endpoint string
	log      *slog.Logger
}

const SYSTEM_PROMPT = `
You are a helpful assistant that can query the meteo PostgreSQL database 
to get information about weather for cities in the world. 

For now just tell the user what you can do.

Here is the user prompt:
{{.UserPrompt}}
`

func (s *Ollama) PreparePrompt(userPrompt string) (string, error) {
	tmpl, err := template.New("").Parse(
		SYSTEM_PROMPT,
	)
	if err != nil {
		return "", err
	}

	data := struct {
		UserPrompt string
	}{
		UserPrompt: userPrompt,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Ollama) Chat(userPrompt string) (string, error) {
	s.log.Info("preparing prompt", "model", s.model, "endpoint", s.endpoint)

	prompt, err := s.PreparePrompt(userPrompt)
	if err != nil {
		s.log.Error("failed to prepare prompt", "error", err)
		return "", err
	}

	reqBody := Request{
		Prompt: prompt,
		Model:  s.model,
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		s.log.Error("failed to marshal request", "error", err)
		return "", err
	}

	s.log.Info("sending request to ollama", "url", fmt.Sprintf("http://%s/api/generate", s.endpoint))

	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/generate", s.endpoint),
		"application/json",
		bytes.NewBuffer(data),
	)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	s.log.Info("received response from ollama", "status", resp.StatusCode)

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding ollama response: %w", err)
	}

	s.log.Info("chat completed", "response_length", len(result.Response))

	return result.Response, nil
}
