package logging_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

func TestFromContext(t *testing.T) {
	t.Run("it returns the default non nil context given no logger is configured", func(t *testing.T) {
		assert.NotNil(t, logging.FromContext(context.Background()))
	})
	t.Run("it returns the configured context given one was already configured", func(t *testing.T) {
		var (
			logger = zap.NewNop()
			ctx    = logging.With(context.Background(), logger)
		)
		assert.Equal(t, logger, logging.FromContext(ctx))
	})
}

func TestWith(t *testing.T) {
	t.Run("it returns the configured context", func(t *testing.T) {
		var (
			logger = zap.NewNop()
			ctx    = logging.With(context.Background(), logger)
		)
		assert.Equal(t, logger, logging.FromContext(ctx))
	})
}

func TestWithOptions(t *testing.T) {
	t.Run("it returns the original context given no options are passed", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, ctx, logging.WithOptions(ctx))
	})
	t.Run("it returns a context with applied options", func(t *testing.T) {
		const (
			logKey = "someKey"
			logVal = "someVal"
		)

		var (
			logger, obs = observer.New(zapcore.InfoLevel)
			ctx         = logging.With(context.Background(), zap.New(logger))
		)

		ctx = logging.WithOptions(ctx, zap.Fields(zap.String(logKey, logVal)))
		logging.FromContext(ctx).Info("test")

		registeredLogs := obs.All()

		require.Len(t, registeredLogs, 1)
		require.NotNil(t, registeredLogs[0].Context, registeredLogs[0].Context[0])

		assert.Equal(t, logKey, registeredLogs[0].Context[0].Key)
		assert.Equal(t, logVal, registeredLogs[0].Context[0].String)
	})
}

func TestInfo(t *testing.T) {
	t.Run("it logs the expected log", func(t *testing.T) {
		const (
			logMessage = "beep boop"
		)

		var (
			logger, obs = observer.New(zapcore.InfoLevel)
			ctx         = logging.With(context.Background(), zap.New(logger))
		)

		logging.Info(ctx, logMessage)

		registeredLogs := obs.All()

		require.Len(t, registeredLogs, 1)
		assert.Equal(t, logMessage, registeredLogs[0].Message)
	})
}

func TestError(t *testing.T) {
	t.Run("it logs the expected log", func(t *testing.T) {
		const (
			logMessage = "oooopsie"
		)

		var (
			logger, obs = observer.New(zapcore.ErrorLevel)
			ctx         = logging.With(context.Background(), zap.New(logger))
		)

		logging.Error(ctx, logMessage)

		registeredLogs := obs.All()

		require.Len(t, registeredLogs, 1)
		assert.Equal(t, logMessage, registeredLogs[0].Message)
	})
}

func TestDebug(t *testing.T) {
	t.Run("it logs the expected log", func(t *testing.T) {
		const (
			logMessage = "print this and that"
		)

		var (
			logger, obs = observer.New(zapcore.DebugLevel)
			ctx         = logging.With(context.Background(), zap.New(logger))
		)

		logging.Debug(ctx, logMessage)

		registeredLogs := obs.All()

		require.Len(t, registeredLogs, 1)
		assert.Equal(t, logMessage, registeredLogs[0].Message)
	})
}

func TestWarn(t *testing.T) {
	t.Run("it logs the expected log", func(t *testing.T) {
		const (
			logMessage = "RabbitMQ is going nuts!"
		)

		var (
			logger, obs = observer.New(zapcore.WarnLevel)
			ctx         = logging.With(context.Background(), zap.New(logger))
		)

		logging.Warn(ctx, logMessage)

		registeredLogs := obs.All()

		require.Len(t, registeredLogs, 1)
		assert.Equal(t, logMessage, registeredLogs[0].Message)
	})
}
