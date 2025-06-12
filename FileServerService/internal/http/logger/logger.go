package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// InitLogger инициализирует глобальный логгер
func InitLogger(isProduction bool) {
	var cfg zap.Config

	if isProduction {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("cannot initialize zap logger: " + err.Error())
	}

	log = logger.Sugar()
}

// Sync очищает буферы логгера
func Sync() {
	_ = log.Sync()
}

// Exportируемые функции логгирования

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}
