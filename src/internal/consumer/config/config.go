package config

import (
	"strconv"

	"github.com/jfgrea27/meteo/internal/queue"
	"github.com/jfgrea27/meteo/internal/utils"
)

type Config struct {
	AppEnv string

	// Required — the consumer cannot run without knowing which providers to use.
	QueueProvider queue.QueueProvider
	QueueName     string
	BlobProvider  string
	BlobBucket    string
	DBProvider    string
	DatabaseURL   string
	DBSchema      string

	// Provider details — not required at startup because only the active
	// provider needs its credentials. Validated when the provider is constructed.
	AWSAccount          string
	AWSRegion           string
	S3UsePathStyle      bool
	SQSWaitTimeSeconds  int32
}

func Load() Config {
	return Config{
		AppEnv: utils.GetEnvVar("APP_ENV", false),

		QueueProvider: queue.QueueProvider(utils.GetEnvVar("QUEUE_PROVIDER", true)),
		QueueName:     utils.GetEnvVar("QUEUE_NAME", true),
		BlobProvider:  utils.GetEnvVar("BLOB_PROVIDER", true),
		BlobBucket:    utils.GetEnvVar("BLOB_BUCKET", true),
		DBProvider:    utils.GetEnvVar("DB_PROVIDER", true),
		DatabaseURL:   utils.GetEnvVar("DATABASE_URL", true),
		DBSchema:      utils.GetEnvVar("DB_SCHEMA", false),

		AWSAccount:         utils.GetEnvVar("AWS_ACCOUNT", false),
		AWSRegion:          utils.GetEnvVar("AWS_REGION", false),
		S3UsePathStyle:     utils.GetEnvVar("S3_USE_PATH_STYLE", false) == "true",
		SQSWaitTimeSeconds: parseWaitTime(utils.GetEnvVar("SQS_WAIT_TIME_SECONDS", false)),
	}
}

func parseWaitTime(s string) int32 {
	if s == "" {
		return 20
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic("SQS_WAIT_TIME_SECONDS must be a valid integer")
	}
	return int32(v)
}
