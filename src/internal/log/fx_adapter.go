package log

import (
	"github.com/rs/zerolog"
	"go.uber.org/fx/fxevent"
	"strings"
)

type FxEventLoggerAdapter struct {
	logger *zerolog.Logger
}

func NewFxEventLoggerAdapter(logger *zerolog.Logger) FxEventLoggerAdapter {
	return FxEventLoggerAdapter{
		logger: logger,
	}
}

func (f FxEventLoggerAdapter) LogEvent(event fxevent.Event) {
	var localLog *zerolog.Event

	switch event.(type) {
	case *fxevent.Provided:
		providedEvent, ok := event.(*fxevent.Provided)
		if ok {
			if providedEvent.Err == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error().Err(providedEvent.Err)
			}

			localLog.Msgf("%s::%s -> %s", providedEvent.ModuleName, providedEvent.ConstructorName, strings.Join(providedEvent.OutputTypeNames, ","))
		}
	case *fxevent.LoggerInitialized:
		loggerInitialized, ok := event.(*fxevent.LoggerInitialized)
		if ok {
			if loggerInitialized.Err == nil {
				localLog = f.logger.Info()
			} else {
				localLog = f.logger.Error().Err(loggerInitialized.Err)
			}

			localLog.Msgf("logger initialized %s", loggerInitialized.ConstructorName)
		}
	case *fxevent.Invoking:
		loggerInvoking, ok := event.(*fxevent.Invoking)
		if ok {
			localLog = f.logger.Info()
			localLog.Msgf("%s::%s invoking", loggerInvoking.ModuleName, loggerInvoking.FunctionName)
		}
	case *fxevent.Invoked:
		invokedEvent, ok := event.(*fxevent.Invoked)
		if ok {
			if invokedEvent.Err == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error()
				localLog = localLog.Err(invokedEvent.Err)
				localLog.Msg(invokedEvent.Trace)
			}

			localLog.Msgf("%s::%s invoked", invokedEvent.ModuleName, invokedEvent.FunctionName)
		}
	case *fxevent.OnStartExecuting:
		StartExecutingEvent, ok := event.(*fxevent.OnStartExecuting)
		if ok {
			localLog = f.logger.Info()
			localLog.Msgf("%s->%s starting", StartExecutingEvent.FunctionName, StartExecutingEvent.CallerName)
		}
	case *fxevent.OnStartExecuted:
		startExecutedEvent, ok := event.(*fxevent.OnStartExecuted)
		if ok {
			if startExecutedEvent.Err == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error()
				localLog = localLog.Err(startExecutedEvent.Err)
			}

			localLog.Msgf(
				"%s->%s %s %s started",
				startExecutedEvent.FunctionName,
				startExecutedEvent.CallerName,
				startExecutedEvent.Method,
				startExecutedEvent.Runtime,
			)
		}
	case *fxevent.Started:
		startedEvent, ok := event.(*fxevent.Started)
		if ok {
			if startedEvent.Err != nil {
				f.logger.Error().Err(startedEvent.Err).Msg("")
			} else {
				f.logger.Info().Msg("application successfully initialized")
			}
		}
	case *fxevent.Stopping:
		stoppingEvent, ok := event.(*fxevent.Stopping)
		if ok {
			f.logger.Warn().Msgf("stopping application... %s", stoppingEvent.Signal.String())
		}
	case *fxevent.Stopped:
		stoppedEvent, ok := event.(*fxevent.Stopped)
		if ok {
			f.logger.Warn().Err(stoppedEvent.Err).Msg("")
		}
	case *fxevent.OnStopExecuting:
		stopExecutingEvent, ok := event.(*fxevent.OnStopExecuting)
		if ok {
			localLog = f.logger.Info()
			localLog.Msgf("%s->%s stopping", stopExecutingEvent.FunctionName, stopExecutingEvent.CallerName)
		}
	case *fxevent.OnStopExecuted:
		stopExecutedEvent, ok := event.(*fxevent.OnStopExecuted)
		if ok {
			if stopExecutedEvent.Err == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error()
				localLog.Err(stopExecutedEvent.Err).Msg("")
			}

			localLog.Msgf(
				"%s->%s %s stopped",
				stopExecutedEvent.FunctionName,
				stopExecutedEvent.CallerName,
				stopExecutedEvent.Runtime,
			)
		}
	case *fxevent.RollingBack:
		rollingBackEvent, ok := event.(*fxevent.RollingBack)
		if ok {
			if rollingBackEvent.StartErr == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error()
				localLog.Err(rollingBackEvent.StartErr).Msg("")
			}

			localLog.Msg("rolling")
		}
	case *fxevent.RolledBack:
		rolledBackEvent, ok := event.(*fxevent.RolledBack)
		if ok {
			if rolledBackEvent.Err == nil {
				localLog = f.logger.Info()

			} else {
				localLog = f.logger.Error()
				localLog.Err(rolledBackEvent.Err).Msg("")
			}

			localLog.Msgf(
				"rolled",
			)
		}
	default:
		f.logger.Error().Msg("unknown event")
	}
}
