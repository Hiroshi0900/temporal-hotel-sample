package workflows

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"temporal-hotel-sample/internal/activities"
)

// TestHotelBookingSagaWorkflow_WithMissCompensation
// ホテル予約Sagaワークフローの包括的テスト
// - ホテルルーム予約、ディナー食材予約、駐車場予約の順次実行
// - 失敗時の補償処理（逆順実行）とリトライ機能のテスト

// testケース
// 正常系:
//   - ホテルルーム、ディナー食材、駐車場の順に成功する、補償アクションは動かない
//
// 準異常系
//   - ディナー食材予約で失敗するが、ホテルルームの補償アクションが1回で成功して整合性を保てる
//   - ディナー食材予約で失敗するが、ホテルルームの補償アクションが2回失敗してから成功して整合性を保てる
//   - 駐車場予約で失敗するが、ホテルルームとディナー食材の補償アクションが成功して整合性を保てる
//   - 駐車場予約で失敗するが、ホテルルームとディナー食材の補償アクションが2回ずつ失敗してから成功する、最終的には整合性を保てる
//
// 異常系:
//   - ホテルルーム予約で失敗して、エラーを返却して終了する
//   - ディナー食材予約で失敗して、ホテルルームの補償アクションが3回失敗して、エラーを返して終了する
//   - 駐車場予約で失敗して、ホテルルームとディナー食材の補償アクションが3回ずつ失敗して、エラーを返して終了する
func TestHotelBookingSagaWorkflow_WithMissCompensation(t *testing.T) {
	tests := map[string]struct {
		request BookingRequest

		mockHotelError  error
		mockHotelTimes  int
		mockHotelResult *activities.HotelBookingResult

		mockDinnerError  error
		mockDinnerTimes  int
		mockDinnerResult *activities.DinnerBookingResult

		mockParkingError  error
		mockParkingTimes  int
		mockParkingResult *activities.ParkingBookingResult

		mockHotelCompensationError  error
		mockHotelCompensationTimes  int
		mockHotelCompensationResult *activities.CompensationResult

		mockDinnerCompensationError  error
		mockDinnerCompensationTimes  int
		mockDinnerCompensationResult *activities.CompensationResult

		expectedWorkflowSuccess bool
		expectedHotelSuccess    bool
		expectedDinnerSuccess   bool
		expectedParkingSuccess  bool
	}{
		// 正常系: ホテルルーム、ディナー食材、駐車場の順に成功する、補償アクションは動かない
		"正常系 - ホテルルーム、ディナー食材、駐車場の順に成功": {
			request: BookingRequest{
				BookingID: "booking-success-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "standard"},
				Parking:   ParkingRequest{SpaceType: "standard"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-001",
				Message:    "ホテルルーム予約が完了しました",
			},
			mockDinnerResult: &activities.DinnerBookingResult{
				Success:    true,
				ResourceID: "food-001",
				Message:    "ディナー食材予約が完了しました",
			},

			mockParkingResult: &activities.ParkingBookingResult{
				Success:    true,
				ResourceID: "parking-001",
				Message:    "駐車場予約が完了しました",
			},

			expectedWorkflowSuccess: true,
			expectedHotelSuccess:    true,
			expectedDinnerSuccess:   true,
			expectedParkingSuccess:  true,
		},
		// 準異常系: ディナー食材予約で失敗するが、ホテルルームの補償アクションが1回で成功
		"準異常系 - ディナー食材予約失敗、ホテルルーム補償1回で成功": {
			request: BookingRequest{
				BookingID: "booking-dinner-fail-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "out-of-stock"},
				Parking:   ParkingRequest{SpaceType: "standard"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-002",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerError: &activities.BusinessError{Message: "指定されたメニューの食材が在庫不足です"},
			mockDinnerTimes: 3,
			mockHotelCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ホテルルーム補償が完了しました",
			},

			expectedWorkflowSuccess: false,
			expectedHotelSuccess:    true,
			expectedDinnerSuccess:   false,
			expectedParkingSuccess:  false,
		},
		// 準異常系: ディナー食材予約で失敗するが、ホテルルームの補償アクションが2回失敗してから成功
		"準異常系 - ディナー食材予約失敗、ホテルルーム補償2回失敗後成功": {
			request: BookingRequest{
				BookingID: "booking-dinner-fail-002",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "out-of-stock"},
				Parking:   ParkingRequest{SpaceType: "standard"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-003",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerError:            &activities.BusinessError{Message: "指定されたメニューの食材が在庫不足です"},
			mockDinnerTimes:            3,
			mockHotelCompensationError: &activities.ServerError{Message: "補償処理で一時的エラーが発生しました"},
			mockHotelCompensationTimes: 2,
			mockHotelCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ホテルルーム補償が完了しました",
			},

			expectedWorkflowSuccess: false,
			expectedHotelSuccess:    true,
			expectedDinnerSuccess:   false,
			expectedParkingSuccess:  false,
		},
		// 準異常系: 駐車場予約で失敗するが、ホテルルーム・ディナー食材の補償アクションが成功
		"準異常系 - 駐車場予約失敗、ホテルルーム・ディナー食材補償成功": {
			request: BookingRequest{
				BookingID: "booking-parking-fail-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "standard"},
				Parking:   ParkingRequest{SpaceType: "full"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-004",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerResult: &activities.DinnerBookingResult{
				Success:    true,
				ResourceID: "food-004",
				Message:    "ディナー食材予約が完了しました",
			},

			mockParkingError: &activities.BusinessError{Message: "指定された駐車場は満車です"},
			mockParkingTimes: 3,
			mockHotelCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ホテルルーム補償が完了しました",
			},

			mockDinnerCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ディナー食材補償が完了しました",
			},

			expectedWorkflowSuccess: false,
			expectedHotelSuccess:    true,
			expectedDinnerSuccess:   true,
			expectedParkingSuccess:  false,
		},
		// 準異常系: 駐車場予約で失敗、ホテルルーム・ディナー食材の補償アクションが2回ずつ失敗してから成功
		"準異常系 - 駐車場予約失敗、ホテルルーム・ディナー食材補償2回失敗後成功": {
			request: BookingRequest{
				BookingID: "booking-parking-fail-002",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "standard"},
				Parking:   ParkingRequest{SpaceType: "full"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-005",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerResult: &activities.DinnerBookingResult{
				Success:    true,
				ResourceID: "food-005",
				Message:    "ディナー食材予約が完了しました",
			},

			mockParkingError:           &activities.BusinessError{Message: "指定された駐車場は満車です"},
			mockParkingTimes:           3,
			mockHotelCompensationError: &activities.ServerError{Message: "ホテル補償処理で一時的エラーが発生しました"},
			mockHotelCompensationTimes: 2,
			mockHotelCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ホテルルーム補償が完了しました",
			},

			mockDinnerCompensationError: &activities.ServerError{Message: "ディナー補償処理で一時的エラーが発生しました"},
			mockDinnerCompensationTimes: 2,
			mockDinnerCompensationResult: &activities.CompensationResult{
				Success: true,
				Message: "ディナー食材補償が完了しました",
			},

			expectedWorkflowSuccess: false,
			expectedHotelSuccess:    true,
			expectedDinnerSuccess:   true,
			expectedParkingSuccess:  false,
		},
		// 異常系: ホテルルーム予約で失敗して、エラーを返却して終了
		"異常系 - ホテルルーム予約失敗、エラー返却終了": {
			request: BookingRequest{
				BookingID: "booking-hotel-fail-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-full"},
				Dinner:    DinnerRequest{MenuType: "standard"},
				Parking:   ParkingRequest{SpaceType: "standard"},
			},
			mockHotelError:          &activities.BusinessError{Message: "指定されたホテルは満室です"},
			mockHotelTimes:          3,
			expectedWorkflowSuccess: false,
			expectedHotelSuccess:    false,
			expectedDinnerSuccess:   false,
			expectedParkingSuccess:  false,
		},
		// 異常系: ディナー食材予約で失敗、ホテルルームの補償アクションが3回失敗してエラー終了
		"異常系 - ディナー食材予約失敗、ホテルルーム補償3回失敗": {
			request: BookingRequest{
				BookingID: "booking-dinner-fail-comp-fail-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "out-of-stock"},
				Parking:   ParkingRequest{SpaceType: "standard"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-006",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerError:            &activities.BusinessError{Message: "指定されたメニューの食材が在庫不足です"},
			mockDinnerTimes:            3,
			mockHotelCompensationError: &activities.ServerError{Message: "補償処理システムがダウンしています"},
			mockHotelCompensationTimes: 3, // リトライ回数上限
			expectedWorkflowSuccess:    false,
			expectedHotelSuccess:       true,
			expectedDinnerSuccess:      false,
			expectedParkingSuccess:     false,
		},
		// 異常系: 駐車場予約で失敗、ホテルルーム・ディナー食材の補償アクションが3回ずつ失敗してエラー終了
		"異常系 - 駐車場予約失敗、ホテルルーム・ディナー食材補償3回失敗": {
			request: BookingRequest{
				BookingID: "booking-parking-fail-comp-fail-001",
				UserID:    "user-001",
				Hotel:     HotelRequest{HotelID: "hotel-001"},
				Dinner:    DinnerRequest{MenuType: "standard"},
				Parking:   ParkingRequest{SpaceType: "full"},
			},
			mockHotelResult: &activities.HotelBookingResult{
				Success:    true,
				ResourceID: "room-007",
				Message:    "ホテルルーム予約が完了しました",
			},

			mockDinnerResult: &activities.DinnerBookingResult{
				Success:    true,
				ResourceID: "food-007",
				Message:    "ディナー食材予約が完了しました",
			},

			mockParkingError:            &activities.BusinessError{Message: "指定された駐車場は満車です"},
			mockParkingTimes:            3,
			mockHotelCompensationError:  &activities.ServerError{Message: "ホテル補償処理システムがダウンしています"},
			mockHotelCompensationTimes:  3, // リトライ回数上限
			mockDinnerCompensationError: &activities.ServerError{Message: "ディナー補償処理システムがダウンしています"},
			mockDinnerCompensationTimes: 3, // リトライ回数上限
			expectedWorkflowSuccess:     false,
			expectedHotelSuccess:        true,
			expectedDinnerSuccess:       true,
			expectedParkingSuccess:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// when - ワークフローを実行
			testSuite := &testsuite.WorkflowTestSuite{}
			testEnv := testSuite.NewTestWorkflowEnvironment()

			// アクティビティの登録
			testEnv.RegisterActivity(activities.HotelRoomBookingActivity)
			testEnv.RegisterActivity(activities.DinnerFoodBookingActivity)
			testEnv.RegisterActivity(activities.ParkingBookingActivity)
			testEnv.RegisterActivity(activities.CompensateHotelRoomActivity)
			testEnv.RegisterActivity(activities.CompensateDinnerFoodActivity)
			testEnv.RegisterActivity(activities.CompensateParkingActivity)

			// モックの設定
			// ホテルルーム予約
			if tt.mockHotelTimes > 0 {
				testEnv.OnActivity(activities.HotelRoomBookingActivity, mock.Anything, mock.Anything).Return(
					nil, tt.mockHotelError).Times(tt.mockHotelTimes)
			}
			if tt.mockHotelResult != nil {
				testEnv.OnActivity(activities.HotelRoomBookingActivity, mock.Anything, mock.Anything).Return(
					tt.mockHotelResult, nil)
			}

			// ディナー食材予約
			if tt.mockDinnerTimes > 0 {
				testEnv.OnActivity(activities.DinnerFoodBookingActivity, mock.Anything, mock.Anything).Return(
					nil, tt.mockDinnerError).Times(tt.mockDinnerTimes)
			}
			if tt.mockDinnerResult != nil {
				testEnv.OnActivity(activities.DinnerFoodBookingActivity, mock.Anything, mock.Anything).Return(
					tt.mockDinnerResult, nil)
			}

			// 駐車場予約
			if tt.mockParkingTimes > 0 {
				testEnv.OnActivity(activities.ParkingBookingActivity, mock.Anything, mock.Anything).Return(
					nil, tt.mockParkingError).Times(tt.mockParkingTimes)
			}
			if tt.mockParkingResult != nil {
				testEnv.OnActivity(activities.ParkingBookingActivity, mock.Anything, mock.Anything).Return(
					tt.mockParkingResult, nil)
			}

			// ホテルルーム補償処理
			if tt.mockHotelCompensationTimes > 0 {
				testEnv.OnActivity(activities.CompensateHotelRoomActivity, mock.Anything, mock.Anything, mock.Anything).Return(
					nil, tt.mockHotelCompensationError).Times(tt.mockHotelCompensationTimes)
			}
			if tt.mockHotelCompensationResult != nil {
				testEnv.OnActivity(activities.CompensateHotelRoomActivity, mock.Anything, mock.Anything, mock.Anything).Return(
					tt.mockHotelCompensationResult, nil)
			}

			// ディナー食材補償処理
			if tt.mockDinnerCompensationTimes > 0 {
				testEnv.OnActivity(activities.CompensateDinnerFoodActivity, mock.Anything, mock.Anything, mock.Anything).Return(
					nil, tt.mockDinnerCompensationError).Times(tt.mockDinnerCompensationTimes)
			}
			if tt.mockDinnerCompensationResult != nil {
				testEnv.OnActivity(activities.CompensateDinnerFoodActivity, mock.Anything, mock.Anything, mock.Anything).Return(
					tt.mockDinnerCompensationResult, nil)
			}

			testEnv.ExecuteWorkflow(HotelBookingSaga, tt.request)

			// then
			if !testEnv.IsWorkflowCompleted() {
				t.Errorf("ワークフローが完了していません")
				return
			}

			// ワークフローはエラーなく完了するが、結果はSuccessがfalse
			err := testEnv.GetWorkflowError()
			if err != nil {
				t.Errorf("ワークフローがエラーで終了しました: %v", err)
				return
			}

			// 結果の取得と検証
			var result BookingResult
			if err := testEnv.GetWorkflowResult(&result); err != nil {
				t.Errorf("結果の取得に失敗しました: %v", err)
				return
			}

			// then
			assert.Equal(t, tt.expectedWorkflowSuccess, result.Success)
			assert.Equal(t, tt.expectedHotelSuccess, result.HotelResult != nil && result.HotelResult.Success)
			assert.Equal(t, tt.expectedDinnerSuccess, result.DinnerResult != nil && result.DinnerResult.Success)
			assert.Equal(t, tt.expectedParkingSuccess, result.ParkingResult != nil && result.ParkingResult.Success)

			// 補償処理の呼び出し確認（リトライを含む）
			testEnv.AssertExpectations(t)
		})
	}
}
