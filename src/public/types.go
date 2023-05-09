package public

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

type RabbitMqLogConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Queue    *string
	Exchange *string
	// ContentType default json
	ContentType *string
	// Timeout default 250 ms
	Timeout *time.Duration
	// LogsChannelSize default 1000
	LogsChannelSize *int
}

type ModuleConfig struct {
	OwnInstance *zerolog.Logger

	LogToConsole  bool
	LogToFile     *os.File
	LogToRabbitMq *RabbitMqLogConfig

	LogOutputCustomWriters []io.Writer
}
