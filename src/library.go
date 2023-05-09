package zerologfx

import (
	"context"
	"github.com/TrueGameover/zerologfx/src/internal/log"
	"github.com/TrueGameover/zerologfx/src/internal/types"
	"github.com/TrueGameover/zerologfx/src/public"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"io"
	"os"
	"sync"
	"time"
)

//goland:noinspection GoUnusedExportedFunction
func NewZerologFxModule(appCtx context.Context, config public.ModuleConfig) fx.Option {
	if config.LogToRabbitMq != nil && config.LogToRabbitMq.Queue == nil && config.LogToRabbitMq.Exchange == nil {
		panic("queue or exchange expected")
	}

	return fx.Module("zerologfx",
		fx.Provide(
			log.NewFxEventLoggerAdapter,
			newZeroLogLogger,
			log.NewZeroLogRabbitMqAdapter,
		),
		fx.Provide(
			fx.Private,
			func() *types.ZeroLogFxModule {
				return &types.ZeroLogFxModule{
					Config: config,
					AppCtx: appCtx,
				}
			},
		),
	)
}

//goland:noinspection GoUnusedExportedFunction
func RegisterLoggerForModule() fx.Option {
	return fx.WithLogger(func(adapter log.FxEventLoggerAdapter) fxevent.Logger {
		return adapter
	})
}

func newZeroLogLogger(
	lf fx.Lifecycle,
	mod *types.ZeroLogFxModule,
	rabbitmqZeroLogAdapter *log.ZeroLogRabbitMqAdapter,
) *zerolog.Logger {
	var appLogger zerolog.Logger

	if mod.Config.OwnInstance != nil {
		appLogger = *mod.Config.OwnInstance
	} else {
		appLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	writers := make([]io.Writer, 0, 3)

	if mod.Config.LogToConsole {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	if mod.Config.LogToFile != nil {
		writers = append(writers, zerolog.ConsoleWriter{Out: mod.Config.LogToFile, TimeFormat: time.RFC3339, NoColor: true})
	}

	if mod.Config.LogToRabbitMq != nil {
		writers = append(writers, rabbitmqZeroLogAdapter)
	}

	if mod.Config.LogOutputCustomWriters != nil {
		writers = append(writers, mod.Config.LogOutputCustomWriters...)
	}

	appLogger = appLogger.Output(
		zerolog.MultiLevelWriter(writers...),
	)

	wGroup := sync.WaitGroup{}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			appLogger.Info().Msg("=================== START ===================")

			if mod.Config.LogToRabbitMq != nil {
				wGroup.Add(1)
				go func() {
					defer wGroup.Done()
					for {
						err := rabbitmqZeroLogAdapter.Handle(mod.AppCtx)
						if err != nil {
							appLogger.Error().Err(err).Msg("zerologfx rabbitmq publish")
						}

						select {
						case <-mod.AppCtx.Done():
							return
						default:
						}
					}
				}()
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			wGroup.Wait()
			return nil
		},
	})

	return &appLogger
}
