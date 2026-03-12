package ai

type Request struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type Response struct {
	Response string `json:"response"`
}
