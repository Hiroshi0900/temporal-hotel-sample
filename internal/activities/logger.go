package activities

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// Logger ロガーインターフェース
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// TemporalLogger Temporal用のロガー実装
type TemporalLogger struct {
	ctx context.Context
}

// NewTemporalLogger Temporal用ロガーのコンストラクタ
func NewTemporalLogger(ctx context.Context) Logger {
	return &TemporalLogger{ctx: ctx}
}

func (t *TemporalLogger) Info(msg string, keysAndValues ...interface{}) {
	logger := activity.GetLogger(t.ctx)
	logger.Info(msg, keysAndValues...)
}

func (t *TemporalLogger) Error(msg string, keysAndValues ...interface{}) {
	logger := activity.GetLogger(t.ctx)
	logger.Error(msg, keysAndValues...)
}
