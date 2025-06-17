package workflows

import (
	"errors"
	"testing"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// テスト用のアクティビティ関数型を定義
type TestActivity func() error

func TestCompensations_AddCompensation(t *testing.T) {
	tests := []struct {
		name     string
		given    struct {
			activities []TestActivity
		}
		when string
		then struct {
			expectedLength int
		}
	}{
		{
			name: "補償処理を1つ追加",
			given: struct {
				activities []TestActivity
			}{[]TestActivity{func() error { return nil }}},
			when: "add_compensation",
			then: struct {
				expectedLength int
			}{1},
		},
		{
			name: "補償処理を複数追加",
			given: struct {
				activities []TestActivity
			}{[]TestActivity{
				func() error { return nil },
				func() error { return nil },
				func() error { return nil },
			}},
			when: "add_compensation",
			then: struct {
				expectedLength int
			}{3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			compensations := Compensations{}

			// When
			for _, activity := range tt.given.activities {
				compensations.AddCompensation(activity)
			}

			// Then
			if len(compensations) != tt.then.expectedLength {
				t.Errorf("期待された長さと異なります。expected: %d, actual: %d", tt.then.expectedLength, len(compensations))
			}
		})
	}
}

func TestCompensations_Compensate_Sequential_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Given
	compensations := Compensations{}
	activity1 := func() error { return nil }
	activity2 := func() error { return nil }
	activity3 := func() error { return nil }

	compensations.AddCompensation(activity1)
	compensations.AddCompensation(activity2)
	compensations.AddCompensation(activity3)

	// アクティビティの登録
	env.RegisterActivity(activity1)
	env.RegisterActivity(activity2)
	env.RegisterActivity(activity3)

	// When & Then
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		compensations.Compensate(ctx, false) // 順次実行
		return nil
	})

	if !env.IsWorkflowCompleted() {
		t.Error("ワークフローが完了していません")
	}

	err := env.GetWorkflowError()
	if err != nil {
		t.Errorf("予期しないエラーが発生しました: %v", err)
	}
}

func TestCompensations_Compensate_Sequential_WithErrors(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Given
	compensations := Compensations{}
	activity1 := func() error { return nil }
	activity2 := func() error { return errors.New("補償処理でエラー発生") }
	activity3 := func() error { return nil }

	compensations.AddCompensation(activity1)
	compensations.AddCompensation(activity2)
	compensations.AddCompensation(activity3)

	// アクティビティの登録
	env.RegisterActivity(activity1)
	env.RegisterActivity(activity2)
	env.RegisterActivity(activity3)

	// When & Then
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		compensations.Compensate(ctx, false) // 順次実行
		return nil
	})

	if !env.IsWorkflowCompleted() {
		t.Error("ワークフローが完了していません")
	}

	// 補償処理でエラーが発生してもワークフロー自体は完了する
	err := env.GetWorkflowError()
	if err != nil {
		t.Errorf("予期しないエラーが発生しました: %v", err)
	}
}

func TestCompensations_Compensate_Parallel_Success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Given
	compensations := Compensations{}
	activity1 := func() error { return nil }
	activity2 := func() error { return nil }
	activity3 := func() error { return nil }

	compensations.AddCompensation(activity1)
	compensations.AddCompensation(activity2)
	compensations.AddCompensation(activity3)

	// アクティビティの登録
	env.RegisterActivity(activity1)
	env.RegisterActivity(activity2)
	env.RegisterActivity(activity3)

	// When & Then
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		compensations.Compensate(ctx, true) // 並列実行
		return nil
	})

	if !env.IsWorkflowCompleted() {
		t.Error("ワークフローが完了していません")
	}

	err := env.GetWorkflowError()
	if err != nil {
		t.Errorf("予期しないエラーが発生しました: %v", err)
	}
}

func TestCompensations_Compensate_Parallel_WithErrors(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Given
	compensations := Compensations{}
	activity1 := func() error { return nil }
	activity2 := func() error { return errors.New("並列補償処理でエラー発生") }
	activity3 := func() error { return nil }

	compensations.AddCompensation(activity1)
	compensations.AddCompensation(activity2)
	compensations.AddCompensation(activity3)

	// アクティビティの登録
	env.RegisterActivity(activity1)
	env.RegisterActivity(activity2)
	env.RegisterActivity(activity3)

	// When & Then
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		compensations.Compensate(ctx, true) // 並列実行
		return nil
	})

	if !env.IsWorkflowCompleted() {
		t.Error("ワークフローが完了していません")
	}

	// 補償処理でエラーが発生してもワークフロー自体は完了する
	err := env.GetWorkflowError()
	if err != nil {
		t.Errorf("予期しないエラーが発生しました: %v", err)
	}
}

func TestCompensations_Compensate_EmptyCompensations(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Given
	compensations := Compensations{} // 空の補償処理

	// When & Then
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		compensations.Compensate(ctx, false) // 何も実行されない
		return nil
	})

	if !env.IsWorkflowCompleted() {
		t.Error("ワークフローが完了していません")
	}

	err := env.GetWorkflowError()
	if err != nil {
		t.Errorf("予期しないエラーが発生しました: %v", err)
	}
}
