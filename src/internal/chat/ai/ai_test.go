package ai

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConstructAiChat_Ollama(t *testing.T) {
	log := slog.Default()
	chat := ConstructAiChat(log, "ollama", "localhost:11434", "llama3")

	if chat == nil {
		t.Fatal("expected non-nil AIChat")
	}

	ollama, ok := chat.(*Ollama)
	if !ok {
		t.Fatal("expected *Ollama type")
	}
	if ollama.model != "llama3" {
		t.Errorf("expected model llama3, got %s", ollama.model)
	}
	if ollama.endpoint != "localhost:11434" {
		t.Errorf("expected endpoint localhost:11434, got %s", ollama.endpoint)
	}
}

func TestConstructAiChat_InvalidProvider(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for invalid provider")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "not a valid ai provider") {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	ConstructAiChat(slog.Default(), "invalid", "endpoint", "model")
}

func TestOllama_PreparePrompt(t *testing.T) {
	o := &Ollama{
		model:    "llama3",
		endpoint: "localhost:11434",
		log:      slog.Default(),
	}

	result, err := o.PreparePrompt("What is the weather in Paris?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "What is the weather in Paris?") {
		t.Error("expected prompt to contain user input")
	}
	if !strings.Contains(result, "meteo PostgreSQL database") {
		t.Error("expected prompt to contain system instructions")
	}
}

func TestOllama_PreparePrompt_HTMLEscaping(t *testing.T) {
	o := &Ollama{
		model:    "llama3",
		endpoint: "localhost:11434",
		log:      slog.Default(),
	}

	result, err := o.PreparePrompt("<script>alert('xss')</script>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// html/template should escape HTML
	if strings.Contains(result, "<script>") {
		t.Error("expected HTML to be escaped in prompt")
	}
}

func TestOllama_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected /api/generate, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Model != "llama3" {
			t.Errorf("expected model llama3, got %s", req.Model)
		}
		if !strings.Contains(req.Prompt, "Hello") {
			t.Error("expected prompt to contain user input")
		}

		resp := Response{Response: "I can help you with weather data!"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Strip the "http://" prefix since Ollama.Chat adds it
	endpoint := strings.TrimPrefix(server.URL, "http://")

	o := &Ollama{
		model:    "llama3",
		endpoint: endpoint,
		log:      slog.Default(),
	}

	result, err := o.Chat("Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "I can help you with weather data!" {
		t.Errorf("unexpected response: %s", result)
	}
}

func TestOllama_Chat_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{Response: ""}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	endpoint := strings.TrimPrefix(server.URL, "http://")
	o := &Ollama{
		model:    "llama3",
		endpoint: endpoint,
		log:      slog.Default(),
	}

	result, err := o.Chat("Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty response, got %s", result)
	}
}
