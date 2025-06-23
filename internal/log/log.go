package log

import (
	"io"

	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger = zap.NewNop()

func Init(loglevel zapcore.Level, output io.Writer) {
	Logger = prettyconsole.NewLogger(loglevel).
		WithOptions(zap.ErrorOutput(zapcore.AddSync(output)))
}

func Close() error {
	_ = Logger.Sync()
	return nil
}
