package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Compensations 補償処理のスライス
type Compensations []interface{}

// AddCompensation 補償処理を追加
func (s *Compensations) AddCompensation(activity interface{}) {
	*s = append(*s, activity)
}

// Compensate 補償処理を実行
// inParallel: true=並列実行, false=順次実行（逆順）
func (s Compensations) Compensate(ctx workflow.Context, inParallel bool) {
	// 補償処理用のActivityOptions
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	if !inParallel {
		// 順次実行（逆順）
		for i := len(s) - 1; i >= 0; i-- {
			errCompensation := workflow.ExecuteActivity(ctx, s[i]).Get(ctx, nil)
			if errCompensation != nil {
				workflow.GetLogger(ctx).Error("Executing compensation failed", "Error", errCompensation)
			}
		}
	} else {
		// 並列実行
		selector := workflow.NewSelector(ctx)
		for i := 0; i < len(s); i++ {
			execution := workflow.ExecuteActivity(ctx, s[i])
			selector.AddFuture(execution, func(f workflow.Future) {
				if errCompensation := f.Get(ctx, nil); errCompensation != nil {
					workflow.GetLogger(ctx).Error("Executing compensation failed", "Error", errCompensation)
				}
			})
		}
		for range s {
			selector.Select(ctx)
		}
	}
}
