package lib

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func init() {
	InitLogger()
}

func InitLogger() {
	Logger, _ = zap.NewProduction()
	SugarLogger = Logger.Sugar()
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
}
