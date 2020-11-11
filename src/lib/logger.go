package lib

import "go.uber.org/zap"

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func InitLogger() {
	Logger, _ = zap.NewProduction()
	SugarLogger = Logger.Sugar()
}

func SyncLogger() {
	Logger.Sync()
	SugarLogger.Sync()
}
