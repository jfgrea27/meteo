package main

import (
	"context"
	"os"

	"github.com/jfgrea27/meteo/internal/blob"
	"github.com/jfgrea27/meteo/internal/consumer/config"
	"github.com/jfgrea27/meteo/internal/consumer/handler"
	"github.com/jfgrea27/meteo/internal/db"
	"github.com/jfgrea27/meteo/internal/logger"
	"github.com/jfgrea27/meteo/internal/queue"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("starting consumer")

	ctx := context.Background()

	// setup queue service
	log.Info("initialising queue service", "provider", string(cfg.QueueProvider), "queue", cfg.QueueName)
	queueSvc := queue.ConstructQueueService(cfg.QueueProvider, cfg.AWSAccount, cfg.AWSRegion, cfg.SQSWaitTimeSeconds)

	// setup blob store
	log.Info("initialising blob store", "provider", cfg.BlobProvider, "bucket", cfg.BlobBucket)
	blobStore := blob.ConstructBlobStore(log, cfg.BlobProvider, cfg.BlobBucket, cfg.AWSRegion, cfg.S3UsePathStyle)

	// setup database
	log.Info("initialising database", "provider", cfg.DBProvider)
	database := db.ConstructDatabase(log, cfg.DBProvider, cfg.DatabaseURL, cfg.DBSchema)
	defer database.Close()

	// setup handler
	h := handler.New(log, blobStore, database)

	log.Info("consuming messages", "queue", cfg.QueueName)
	if err := queueSvc.Consume(ctx, h.Handle, cfg.QueueName); err != nil {
		log.Error("consumer error", "error", err)
		os.Exit(1)
	}
}
