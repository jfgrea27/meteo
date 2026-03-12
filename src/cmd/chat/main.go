package main

import (
	"fmt"
	"net"

	"github.com/jfgrea27/rt-meteo/internal/chat/ai"
	"github.com/jfgrea27/rt-meteo/internal/chat/config"
	"github.com/jfgrea27/rt-meteo/internal/chat/handler"
	"github.com/jfgrea27/rt-meteo/internal/logger"
	"google.golang.org/grpc"

	pb "github.com/jfgrea27/rt-meteo/proto"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("starting chat server", "port", cfg.Port)

	// database := db.ConstructDatabase(log, cfg.DBProvider, cfg.DatabaseURL, cfg.DBSchema)
	// defer database.Close()

	// Listen on TCP port
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Error("failed to listen", "error", err)
		return
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Register our service implementation
	aiChat := ai.ConstructAiChat(log, cfg.AIProvider, cfg.AIEndpoint, cfg.AIModel)
	pb.RegisterWeatherChatServer(s, &handler.WeatherChatHandler{AI: aiChat})

	log.Info(fmt.Sprintf("Server listening on %s", cfg.Port))
	if err := s.Serve(lis); err != nil {
		log.Error("failed to serve", "error", err)
	}
}
