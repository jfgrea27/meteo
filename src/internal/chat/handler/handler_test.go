package handler

import (
	"context"
	"errors"
	"io"
	"testing"

	pb "github.com/jfgrea27/meteo/proto"
	"google.golang.org/grpc/metadata"
)

// mockAIChat implements ai.AIChat for testing
type mockAIChat struct {
	response string
	err      error
}

func (m *mockAIChat) PreparePrompt(input string) (string, error) {
	return input, nil
}

func (m *mockAIChat) Chat(input string) (string, error) {
	return m.response, m.err
}

// mockStream implements grpc.BidiStreamingServer[pb.ChatRequest, pb.ChatResponse]
type mockStream struct {
	requests  []*pb.ChatRequest
	responses []*pb.ChatResponse
	recvIdx   int
	sendErr   error
	ctx       context.Context
}

func (m *mockStream) Send(resp *pb.ChatResponse) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.responses = append(m.responses, resp)
	return nil
}

func (m *mockStream) Recv() (*pb.ChatRequest, error) {
	if m.recvIdx >= len(m.requests) {
		return nil, io.EOF
	}
	req := m.requests[m.recvIdx]
	m.recvIdx++
	return req, nil
}

func (m *mockStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockStream) SendHeader(metadata.MD) error  { return nil }
func (m *mockStream) SetTrailer(metadata.MD)         {}
func (m *mockStream) Context() context.Context       { return m.ctx }
func (m *mockStream) SendMsg(interface{}) error      { return nil }
func (m *mockStream) RecvMsg(interface{}) error      { return nil }

func newMockStream(requests []*pb.ChatRequest) *mockStream {
	return &mockStream{
		requests: requests,
		ctx:      context.Background(),
	}
}

func TestChat_SingleMessage(t *testing.T) {
	mock := &mockAIChat{response: "sunny weather today"}
	handler := &WeatherChatHandler{AI: mock}

	stream := newMockStream([]*pb.ChatRequest{
		{Text: "what's the weather?"},
	})

	err := handler.Chat(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(stream.responses))
	}
	if stream.responses[0].Text != "sunny weather today" {
		t.Errorf("expected 'sunny weather today', got '%s'", stream.responses[0].Text)
	}
}

func TestChat_MultipleMessages(t *testing.T) {
	mock := &mockAIChat{response: "response"}
	handler := &WeatherChatHandler{AI: mock}

	stream := newMockStream([]*pb.ChatRequest{
		{Text: "first"},
		{Text: "second"},
		{Text: "third"},
	})

	err := handler.Chat(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(stream.responses))
	}
}

func TestChat_EmptyStream(t *testing.T) {
	mock := &mockAIChat{response: "response"}
	handler := &WeatherChatHandler{AI: mock}

	stream := newMockStream([]*pb.ChatRequest{})

	err := handler.Chat(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.responses) != 0 {
		t.Errorf("expected 0 responses, got %d", len(stream.responses))
	}
}

func TestChat_AIError(t *testing.T) {
	mock := &mockAIChat{err: errors.New("AI service unavailable")}
	handler := &WeatherChatHandler{AI: mock}

	stream := newMockStream([]*pb.ChatRequest{
		{Text: "hello"},
	})

	err := handler.Chat(stream)
	if err == nil {
		t.Fatal("expected error from AI failure")
	}
	if err.Error() != "AI service unavailable" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_SendError(t *testing.T) {
	mock := &mockAIChat{response: "response"}
	handler := &WeatherChatHandler{AI: mock}

	stream := &mockStream{
		requests: []*pb.ChatRequest{
			{Text: "hello"},
		},
		sendErr: errors.New("stream send failed"),
		ctx:     context.Background(),
	}

	err := handler.Chat(stream)
	if err == nil {
		t.Fatal("expected error from stream send failure")
	}
	if err.Error() != "stream send failed" {
		t.Errorf("unexpected error: %v", err)
	}
}
