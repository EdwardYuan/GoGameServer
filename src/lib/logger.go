package lib

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func InitLogger(serviceName string) {
	Logger, _ = zap.NewProduction()
	fileName := "log/" + serviceName + "_" + FormatDateTime(TimeFormat9, time.Now()) + ".log"
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	Logger, err := zap.Config{
		Level:         zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:      "json",
		EncoderConfig: cfg,
		OutputPaths:   []string{"stdout", fileName},
	}.Build()
	if err != nil {
		log.Fatal("InitLogger Error:", err)
	}
	SugarLogger = Logger.Sugar()
	Logger.WithOptions()
}

func SyncLogger() {
	err := Logger.Sync()
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
	err = SugarLogger.Sync()
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
}

// SysLoggerFatal 使用go自带的log记录fatal
func SysLoggerFatal(err error, msg string) {
	if err != nil {
		log.Fatalf("Fatal: %s: %s", err, msg)
	}
}

// FatalOnError 记录错误并panic，严重错误导致程序无法正常运转时使用
func FatalOnError(err error, msg string) {
	if err != nil {
		Logger.Fatal(msg)
		panic(msg)
	}
}

func LogIfError(err error, msg string) {
	if err != nil {
		Log(zap.ErrorLevel, msg, err)
	}
}

func LogErrorAndReturn(err error, msg string) (isErr bool) {
	if err != nil {
		Log(zap.ErrorLevel, msg, err)
		return true
	}
	return false
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
