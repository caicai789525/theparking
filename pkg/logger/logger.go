// pkg/logger/logger.go
package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger // 导出全局日志实例

func InitLogger(env string) {
	var err error
	if env == "production" {
		Log, err = zap.NewProduction()
	} else {
		Log, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
}

func Sync() {
	_ = Log.Sync()
}
