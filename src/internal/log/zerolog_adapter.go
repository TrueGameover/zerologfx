package log

import (
	"context"
	"errors"
	"fmt"
	"github.com/TrueGameover/zerologfx/src/internal/types"
	"github.com/TrueGameover/zerologfx/src/public"
	"github.com/rabbitmq/amqp091-go"
	"strings"
	"time"
)

type ZeroLogRabbitMqAdapter struct {
	logsChannel chan string
	config      public.RabbitMqLogConfig
}

func NewZeroLogRabbitMqAdapter(mod *types.ZeroLogFxModule) (*ZeroLogRabbitMqAdapter, error) {
	s := 1000
	if mod.Config.LogToRabbitMq != nil && mod.Config.LogToRabbitMq.LogsChannelSize != nil {
		s = *mod.Config.LogToRabbitMq.LogsChannelSize
	}

	return &ZeroLogRabbitMqAdapter{
		logsChannel: make(chan string, s),
		config:      *mod.Config.LogToRabbitMq,
	}, nil
}

func (z *ZeroLogRabbitMqAdapter) Write(p []byte) (n int, err error) {
	// zerolog buffer with message can be changed
	msg := strings.Clone(string(p))

	select {
	case z.logsChannel <- msg:
	default:
		return 0, nil
	}

	return len(p), nil
}

func (z *ZeroLogRabbitMqAdapter) Handle(ctx context.Context) error {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", z.config.User, z.config.Password, z.config.Host, z.config.Port)
	amqpConn, err := amqp091.Dial(dsn)
	if err != nil {
		return err
	}
	defer func(amqpConn *amqp091.Connection) {
		_ = amqpConn.Close()
	}(amqpConn)

	channel, err := amqpConn.Channel()
	if err != nil {
		return err
	}
	defer func(channel *amqp091.Channel) {
		_ = channel.Close()
	}(channel)

	exchange := ""
	if z.config.Exchange != nil {
		exchange = *z.config.Exchange
	}

	queueKey := ""
	if z.config.Queue != nil {
		queueKey = *z.config.Queue
	}

	contentType := "application/json"
	if z.config.ContentType != nil {
		contentType = *z.config.ContentType
	}

	publishTimeout := time.Millisecond * 250
	if z.config.Timeout != nil {
		publishTimeout = *z.config.Timeout
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-z.logsChannel:
			if !ok {
				return errors.New("channel was closed")
			}

			if len(msg) > 0 {
				timeoutCtx, cancelFunc := context.WithTimeout(ctx, publishTimeout)

				err = channel.PublishWithContext(
					timeoutCtx,
					exchange,
					queueKey,
					false,
					false,
					amqp091.Publishing{
						ContentType: contentType,
						Body:        []byte(msg),
					},
				)
				cancelFunc()
				if err != nil {
					return err
				}
			}
		}
	}
}
