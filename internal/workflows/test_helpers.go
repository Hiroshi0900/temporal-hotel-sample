package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"temporal-hotel-sample/internal/activities"
)

// TestScenario テストシナリオの定義
type TestScenario struct {
	Name    string
	Request BookingRequest

	HotelMock              ActivityMock
	DinnerMock             ActivityMock
	ParkingMock            ActivityMock
	HotelCompensationMock  ActivityMock
	DinnerCompensationMock ActivityMock

	Expectations ScenarioExpectations
}

// ActivityMock アクティビティのモック定義
type ActivityMock struct {
	ErrorTimes  int
	Error       error
	ResultTimes int
	Result      interface{}
}

// ScenarioExpectations 期待値の定義
type ScenarioExpectations struct {
	WorkflowSuccess bool
	HotelSuccess    bool
	DinnerSuccess   bool
	ParkingSuccess  bool
}

// TestScenarioBuilder テストシナリオビルダー
type TestScenarioBuilder struct {
	scenario *TestScenario
}

// NewScenarioBuilder 新しいシナリオビルダーを作成
func NewScenarioBuilder(name string) *TestScenarioBuilder {
	return &TestScenarioBuilder{
		scenario: &TestScenario{
			Name: name,
		},
	}
}

// WithBookingRequest 予約リクエストを設定
func (b *TestScenarioBuilder) WithBookingRequest(bookingID, userID string) *TestScenarioBuilder {
	b.scenario.Request = BookingRequest{
		BookingID: bookingID,
		UserID:    userID,
		Hotel:     HotelRequest{HotelID: "hotel-001"},
		Dinner:    DinnerRequest{MenuType: "standard"},
		Parking:   ParkingRequest{SpaceType: "standard"},
	}
	return b
}

// WithSuccessfulHotel 成功するホテル予約を設定
func (b *TestScenarioBuilder) WithSuccessfulHotel(resourceID string) *TestScenarioBuilder {
	b.scenario.HotelMock = ActivityMock{
		ResultTimes: 1,
		Result: &activities.HotelBookingResult{
			Success:    true,
			ResourceID: resourceID,
			Message:    "ホテルルーム予約が完了しました",
		},
	}
	return b
}

// WithFailingHotel 失敗するホテル予約を設定
func (b *TestScenarioBuilder) WithFailingHotel(err error, times int) *TestScenarioBuilder {
	b.scenario.HotelMock = ActivityMock{
		ErrorTimes: times,
		Error:      err,
	}
	return b
}

// WithSuccessfulDinner 成功するディナー予約を設定
func (b *TestScenarioBuilder) WithSuccessfulDinner(resourceID string) *TestScenarioBuilder {
	b.scenario.DinnerMock = ActivityMock{
		ResultTimes: 1,
		Result: &activities.DinnerBookingResult{
			Success:    true,
			ResourceID: resourceID,
			Message:    "ディナー食材予約が完了しました",
		},
	}
	return b
}

// WithFailingDinner 失敗するディナー予約を設定
func (b *TestScenarioBuilder) WithFailingDinner(err error, times int) *TestScenarioBuilder {
	b.scenario.DinnerMock = ActivityMock{
		ErrorTimes: times,
		Error:      err,
	}
	return b
}

// WithSuccessfulParking 成功する駐車場予約を設定
func (b *TestScenarioBuilder) WithSuccessfulParking(resourceID string) *TestScenarioBuilder {
	b.scenario.ParkingMock = ActivityMock{
		ResultTimes: 1,
		Result: &activities.ParkingBookingResult{
			Success:    true,
			ResourceID: resourceID,
			Message:    "駐車場予約が完了しました",
		},
	}
	return b
}

// WithFailingParking 失敗する駐車場予約を設定
func (b *TestScenarioBuilder) WithFailingParking(err error, times int) *TestScenarioBuilder {
	b.scenario.ParkingMock = ActivityMock{
		ErrorTimes: times,
		Error:      err,
	}
	return b
}

// WithSuccessfulHotelCompensation 成功するホテル補償を設定
func (b *TestScenarioBuilder) WithSuccessfulHotelCompensation(times int) *TestScenarioBuilder {
	b.scenario.HotelCompensationMock = ActivityMock{
		ResultTimes: times,
		Result: &activities.CompensationResult{
			Success: true,
			Message: "ホテルルーム補償が完了しました",
		},
	}
	return b
}

// WithFailingHotelCompensation 失敗するホテル補償を設定
func (b *TestScenarioBuilder) WithFailingHotelCompensation(err error, errorTimes int, successTimes int) *TestScenarioBuilder {
	b.scenario.HotelCompensationMock = ActivityMock{
		ErrorTimes:  errorTimes,
		Error:       err,
		ResultTimes: successTimes,
		Result: &activities.CompensationResult{
			Success: true,
			Message: "ホテルルーム補償が完了しました",
		},
	}
	return b
}

// WithSuccessfulDinnerCompensation 成功するディナー補償を設定
func (b *TestScenarioBuilder) WithSuccessfulDinnerCompensation(times int) *TestScenarioBuilder {
	b.scenario.DinnerCompensationMock = ActivityMock{
		ResultTimes: times,
		Result: &activities.CompensationResult{
			Success: true,
			Message: "ディナー食材補償が完了しました",
		},
	}
	return b
}

// WithFailingDinnerCompensation 失敗するディナー補償を設定
func (b *TestScenarioBuilder) WithFailingDinnerCompensation(err error, errorTimes int, successTimes int) *TestScenarioBuilder {
	b.scenario.DinnerCompensationMock = ActivityMock{
		ErrorTimes:  errorTimes,
		Error:       err,
		ResultTimes: successTimes,
		Result: &activities.CompensationResult{
			Success: true,
			Message: "ディナー食材補償が完了しました",
		},
	}
	return b
}

// ExpectAllSuccess 全て成功を期待
func (b *TestScenarioBuilder) ExpectAllSuccess() *TestScenarioBuilder {
	b.scenario.Expectations = ScenarioExpectations{
		WorkflowSuccess: true,
		HotelSuccess:    true,
		DinnerSuccess:   true,
		ParkingSuccess:  true,
	}
	return b
}

// ExpectWorkflowFailure ワークフロー失敗を期待
func (b *TestScenarioBuilder) ExpectWorkflowFailure() *TestScenarioBuilder {
	b.scenario.Expectations.WorkflowSuccess = false
	return b
}

// ExpectHotelSuccess ホテル成功を期待
func (b *TestScenarioBuilder) ExpectHotelSuccess() *TestScenarioBuilder {
	b.scenario.Expectations.HotelSuccess = true
	return b
}

// ExpectHotelFailure ホテル失敗を期待
func (b *TestScenarioBuilder) ExpectHotelFailure() *TestScenarioBuilder {
	b.scenario.Expectations.HotelSuccess = false
	return b
}

// ExpectDinnerSuccess ディナー成功を期待
func (b *TestScenarioBuilder) ExpectDinnerSuccess() *TestScenarioBuilder {
	b.scenario.Expectations.DinnerSuccess = true
	return b
}

// ExpectDinnerFailure ディナー失敗を期待
func (b *TestScenarioBuilder) ExpectDinnerFailure() *TestScenarioBuilder {
	b.scenario.Expectations.DinnerSuccess = false
	return b
}

// ExpectParkingSuccess 駐車場成功を期待
func (b *TestScenarioBuilder) ExpectParkingSuccess() *TestScenarioBuilder {
	b.scenario.Expectations.ParkingSuccess = true
	return b
}

// ExpectParkingFailure 駐車場失敗を期待
func (b *TestScenarioBuilder) ExpectParkingFailure() *TestScenarioBuilder {
	b.scenario.Expectations.ParkingSuccess = false
	return b
}

// Build シナリオを構築
func (b *TestScenarioBuilder) Build() TestScenario {
	return *b.scenario
}

// WorkflowTestHelper ワークフローテストヘルパー
type WorkflowTestHelper struct {
	testEnv *testsuite.TestWorkflowEnvironment
}

// NewWorkflowTestHelper 新しいワークフローテストヘルパーを作成
func NewWorkflowTestHelper() *WorkflowTestHelper {
	testSuite := &testsuite.WorkflowTestSuite{}
	testEnv := testSuite.NewTestWorkflowEnvironment()

	// アクティビティの登録
	testEnv.RegisterActivity(activities.HotelRoomBookingActivity)
	testEnv.RegisterActivity(activities.DinnerFoodBookingActivity)
	testEnv.RegisterActivity(activities.ParkingBookingActivity)
	testEnv.RegisterActivity(activities.CompensateHotelRoomActivity)
	testEnv.RegisterActivity(activities.CompensateDinnerFoodActivity)
	testEnv.RegisterActivity(activities.CompensateParkingActivity)

	return &WorkflowTestHelper{
		testEnv: testEnv,
	}
}

// SetupMocks モックを設定
func (h *WorkflowTestHelper) SetupMocks(scenario TestScenario) {
	// ホテル予約モック
	if scenario.HotelMock.ErrorTimes > 0 {
		h.testEnv.OnActivity(activities.HotelRoomBookingActivity, mock.Anything, mock.Anything).Return(
			nil, scenario.HotelMock.Error).Times(scenario.HotelMock.ErrorTimes)
	}
	if scenario.HotelMock.ResultTimes > 0 {
		h.testEnv.OnActivity(activities.HotelRoomBookingActivity, mock.Anything, mock.Anything).Return(
			scenario.HotelMock.Result, nil).Times(scenario.HotelMock.ResultTimes)
	}

	// ディナー予約モック
	if scenario.DinnerMock.ErrorTimes > 0 {
		h.testEnv.OnActivity(activities.DinnerFoodBookingActivity, mock.Anything, mock.Anything).Return(
			nil, scenario.DinnerMock.Error).Times(scenario.DinnerMock.ErrorTimes)
	}
	if scenario.DinnerMock.ResultTimes > 0 {
		h.testEnv.OnActivity(activities.DinnerFoodBookingActivity, mock.Anything, mock.Anything).Return(
			scenario.DinnerMock.Result, nil).Times(scenario.DinnerMock.ResultTimes)
	}

	// 駐車場予約モック
	if scenario.ParkingMock.ErrorTimes > 0 {
		h.testEnv.OnActivity(activities.ParkingBookingActivity, mock.Anything, mock.Anything).Return(
			nil, scenario.ParkingMock.Error).Times(scenario.ParkingMock.ErrorTimes)
	}
	if scenario.ParkingMock.ResultTimes > 0 {
		h.testEnv.OnActivity(activities.ParkingBookingActivity, mock.Anything, mock.Anything).Return(
			scenario.ParkingMock.Result, nil).Times(scenario.ParkingMock.ResultTimes)
	}

	// ホテル補償モック
	if scenario.HotelCompensationMock.ErrorTimes > 0 {
		h.testEnv.OnActivity(activities.CompensateHotelRoomActivity, mock.Anything, mock.Anything, mock.Anything).Return(
			nil, scenario.HotelCompensationMock.Error).Times(scenario.HotelCompensationMock.ErrorTimes)
	}
	if scenario.HotelCompensationMock.ResultTimes > 0 {
		h.testEnv.OnActivity(activities.CompensateHotelRoomActivity, mock.Anything, mock.Anything, mock.Anything).Return(
			scenario.HotelCompensationMock.Result, nil).Times(scenario.HotelCompensationMock.ResultTimes)
	}

	// ディナー補償モック
	if scenario.DinnerCompensationMock.ErrorTimes > 0 {
		h.testEnv.OnActivity(activities.CompensateDinnerFoodActivity, mock.Anything, mock.Anything, mock.Anything).Return(
			nil, scenario.DinnerCompensationMock.Error).Times(scenario.DinnerCompensationMock.ErrorTimes)
	}
	if scenario.DinnerCompensationMock.ResultTimes > 0 {
		h.testEnv.OnActivity(activities.CompensateDinnerFoodActivity, mock.Anything, mock.Anything, mock.Anything).Return(
			scenario.DinnerCompensationMock.Result, nil).Times(scenario.DinnerCompensationMock.ResultTimes)
	}
}

// ExecuteWorkflow ワークフローを実行
func (h *WorkflowTestHelper) ExecuteWorkflow(request BookingRequest) (BookingResult, error) {
	h.testEnv.ExecuteWorkflow(HotelBookingSaga, request)

	if !h.testEnv.IsWorkflowCompleted() {
		return BookingResult{}, assert.AnError
	}

	if err := h.testEnv.GetWorkflowError(); err != nil {
		return BookingResult{}, err
	}

	var result BookingResult
	if err := h.testEnv.GetWorkflowResult(&result); err != nil {
		return BookingResult{}, err
	}

	return result, nil
}

// AssertExpectations モックの期待値をアサート
func (h *WorkflowTestHelper) AssertExpectations(t *testing.T) {
	h.testEnv.AssertExpectations(t)
}

// WorkflowResultAssertions ワークフロー結果のアサーション
type WorkflowResultAssertions struct {
	t      *testing.T
	result BookingResult
}

// NewWorkflowResultAssertions 新しいアサーションヘルパーを作成
func NewWorkflowResultAssertions(t *testing.T, result BookingResult) *WorkflowResultAssertions {
	return &WorkflowResultAssertions{
		t:      t,
		result: result,
	}
}

// AssertWorkflowSuccess ワークフロー成功をアサート
func (a *WorkflowResultAssertions) AssertWorkflowSuccess(expected bool) *WorkflowResultAssertions {
	assert.Equal(a.t, expected, a.result.Success, "ワークフロー成功フラグが期待値と異なります")
	return a
}

// AssertHotelSuccess ホテル成功をアサート
func (a *WorkflowResultAssertions) AssertHotelSuccess(expected bool) *WorkflowResultAssertions {
	if expected {
		assert.NotNil(a.t, a.result.HotelResult, "ホテル結果がnilです")
		if a.result.HotelResult != nil {
			assert.True(a.t, a.result.HotelResult.Success, "ホテル予約が失敗しています")
		}
	} else {
		if a.result.HotelResult != nil {
			assert.False(a.t, a.result.HotelResult.Success, "ホテル予約が成功していますが、失敗が期待されます")
		}
	}
	return a
}

// AssertDinnerSuccess ディナー成功をアサート
func (a *WorkflowResultAssertions) AssertDinnerSuccess(expected bool) *WorkflowResultAssertions {
	if expected {
		assert.NotNil(a.t, a.result.DinnerResult, "ディナー結果がnilです")
		if a.result.DinnerResult != nil {
			assert.True(a.t, a.result.DinnerResult.Success, "ディナー予約が失敗しています")
		}
	} else {
		if a.result.DinnerResult != nil {
			assert.False(a.t, a.result.DinnerResult.Success, "ディナー予約が成功していますが、失敗が期待されます")
		}
	}
	return a
}

// AssertParkingSuccess 駐車場成功をアサート
func (a *WorkflowResultAssertions) AssertParkingSuccess(expected bool) *WorkflowResultAssertions {
	if expected {
		assert.NotNil(a.t, a.result.ParkingResult, "駐車場結果がnilです")
		if a.result.ParkingResult != nil {
			assert.True(a.t, a.result.ParkingResult.Success, "駐車場予約が失敗しています")
		}
	} else {
		if a.result.ParkingResult != nil {
			assert.False(a.t, a.result.ParkingResult.Success, "駐車場予約が成功していますが、失敗が期待されます")
		}
	}
	return a
}

// AssertAllExpectations 全ての期待値をアサート
func (a *WorkflowResultAssertions) AssertAllExpectations(expectations ScenarioExpectations) {
	a.AssertWorkflowSuccess(expectations.WorkflowSuccess).
		AssertHotelSuccess(expectations.HotelSuccess).
		AssertDinnerSuccess(expectations.DinnerSuccess).
		AssertParkingSuccess(expectations.ParkingSuccess)
}

// よく使用されるエラー定義
var (
	OutOfStockError       = &activities.BusinessError{Message: "指定されたメニューの食材が在庫不足です"}
	HotelFullError        = &activities.BusinessError{Message: "指定されたホテルは満室です"}
	ParkingFullError      = &activities.BusinessError{Message: "指定された駐車場は満車です"}
	CompensationError     = &activities.ServerError{Message: "補償処理で一時的エラーが発生しました"}
	SystemDownError       = &activities.ServerError{Message: "補償処理システムがダウンしています"}
)