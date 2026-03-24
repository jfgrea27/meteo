package handler

import (
	"io"
	"log"

	"github.com/jfgrea27/meteo/internal/chat/ai"
	pb "github.com/jfgrea27/meteo/proto"
	"google.golang.org/grpc"
)

type WeatherChatHandler struct {
	pb.UnimplementedWeatherChatServer
	AI ai.AIChat
}

func (h *WeatherChatHandler) Chat(stream grpc.BidiStreamingServer[pb.ChatRequest, pb.ChatResponse]) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		text := req.GetText()
		log.Printf("received chat message: %s", text)

		reply, err := h.AI.Chat(text)
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.ChatResponse{Text: reply}); err != nil {
			return err
		}
	}
}
