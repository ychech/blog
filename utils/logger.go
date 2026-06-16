package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例（zap SugaredLogger），在项目初始化后可用。
var Logger *zap.SugaredLogger

// InitLogger 初始化 zap 日志
func InitLogger() error {
	// 开发环境配置：彩色输出、显示调用位置
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	Logger = logger.Sugar()
	return nil
}
