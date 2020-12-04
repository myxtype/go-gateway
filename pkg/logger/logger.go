package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

const (
	level = zap.DebugLevel // 打印日志的等级
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func init() {
	writer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(newEncoderConfig()),
		writer,
		level,
	)

	Logger = zap.New(core, zap.AddCaller())
	Sugar = Logger.Sugar()
}
