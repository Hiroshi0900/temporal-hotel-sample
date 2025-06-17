package config

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	// TaskQueue タスクキュー名
	TaskQueue = "HOTEL_BOOKING_TASK_QUEUE"
)

// GetRetryPolicy リトライポリシーを取得
func GetRetryPolicy() *temporal.RetryPolicy {
	return &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
		NonRetryableErrorTypes: []string{
			"BusinessError",
		},
	}
}

// GetActivityOptions アクティビティオプションを取得
func GetActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy:         GetRetryPolicy(),
	}
}
