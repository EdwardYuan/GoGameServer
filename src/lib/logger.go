package lib

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func InitLogger(serviceName string) {
	Logger, _ = zap.NewProduction()
	fileName := serviceName + "_" + FormatDateTime(TimeFormat8, time.Now()) + ".log"
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	Logger, _ := zap.Config{
		Level:         zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:      "json",
		EncoderConfig: cfg,
		OutputPaths:   []string{"stdout", fileName},
	}.Build()
	SugarLogger = Logger.Sugar()
	Logger.WithOptions()
}

func SyncLogger() {
	Logger.Sync()
	SugarLogger.Sync()
}

func SysLoggerFatal(err error, msg string) {
	if err != nil {
		log.Fatalf("Fatal: %s: %s", err, msg)
	}
}

func FatalOnError(err error, msg string) {
	if err != nil {
		Logger.Fatal(msg)
		panic(nil)
	}
}

func Log(logLevel zapcore.Level, msg string, err error) {
	if err != nil {
		SugarLogger.Errorf(msg, err)
		return
	}
	switch logLevel {
	case zap.DebugLevel:
		SugarLogger.Debugf(msg)
	case zap.InfoLevel:
		SugarLogger.Infof(msg)
	default:
		SugarLogger.Debugf(msg)
	}
	SyncLogger()
}
