package util

import (
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// GlobalLogHook 全局日志收集器，用于 HTTP API 查询
	GlobalLogHook *MemoryLogHook
)

func InitLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return frame.Function, "" // TODO: 生成包的简写
		},
	})
	logrus.SetReportCaller(true)

	// 创建并添加内存日志收集 hook（默认保存 1000 条日志）
	GlobalLogHook = NewMemoryLogHook(1000)
	logrus.AddHook(GlobalLogHook)
}
