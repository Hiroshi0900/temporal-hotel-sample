package activities

import (
	"strings"
	"testing"

	"go.temporal.io/sdk/testsuite"
)

func TestHotelRoomBookingActivity(t *testing.T) {
	//TODO: テスト構造体は見直す。whenの意味がなくなってる
	tests := []struct {
		name  string
		given HotelBookingRequest
		when  string
		then  struct {
			expectError bool
			errorType   string
			result      *HotelBookingResult
		}
	}{
		{
			name: "正常なホテルルーム予約",
			given: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "execute_activity",
			then: struct {
				expectError bool
				errorType   string
				result      *HotelBookingResult
			}{
				expectError: false,
				errorType:   "",
				result: &HotelBookingResult{
					Success:    true,
					ResourceID: "room-123",
					Message:    "ホテルルーム予約が完了しました",
				},
			},
		},
		{
			name: "BookingIDが空",
			given: HotelBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "execute_activity",
			then: struct {
				expectError bool
				errorType   string
				result      *HotelBookingResult
			}{
				expectError: true,
				errorType:   "BusinessError",
				result:      nil,
			},
		},
		{
			name: "一時的エラー（ネットワークエラー）",
			given: HotelBookingRequest{
				BookingID: "booking-network-error",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "execute_activity",
			then: struct {
				expectError bool
				errorType   string
				result      *HotelBookingResult
			}{
				expectError: true,
				errorType:   "TemporalError",
				result:      nil,
			},
		},
		{
			name: "ビジネスエラー（満室）",
			given: HotelBookingRequest{
				BookingID: "booking-full",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "execute_activity",
			then: struct {
				expectError bool
				errorType   string
				result      *HotelBookingResult
			}{
				expectError: true,
				errorType:   "BusinessError",
				result:      nil,
			},
		},
		{
			name: "冪等性テスト（重複リクエスト）",
			given: HotelBookingRequest{
				BookingID: "booking-duplicate",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "execute_activity",
			then: struct {
				expectError bool
				errorType   string
				result      *HotelBookingResult
			}{
				expectError: false,
				errorType:   "",
				result: &HotelBookingResult{
					Success:    true,
					ResourceID: "room-duplicate",
					Message:    "既に予約済みです",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			// アクティビティを登録
			env.RegisterActivity(HotelRoomBookingActivity)

			// When
			result, err := env.ExecuteActivity(HotelRoomBookingActivity, tt.given)

			// Then
			if tt.then.expectError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
					return
				}

				// Temporalのテスト環境では、エラーはActivityErrorにラップされるため、
				// エラーメッセージをチェックする
				errorMsg := err.Error()
				switch tt.then.errorType {
				case "BusinessError":
					if !strings.Contains(errorMsg, "BookingID is required") &&
						!strings.Contains(errorMsg, "UserID is required") &&
						!strings.Contains(errorMsg, "HotelID is required") &&
						!strings.Contains(errorMsg, "指定されたホテルは満室です") {
						t.Errorf("期待されたBusinessErrorではありません: %s", errorMsg)
					}
				case "TemporalError":
					if !strings.Contains(errorMsg, "ネットワークエラーが発生しました") {
						t.Errorf("期待されたTemporalErrorではありません: %s", errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var actualResult HotelBookingResult
				if result != nil {
					err = result.Get(&actualResult)
					if err != nil {
						t.Errorf("結果の取得に失敗しました: %v", err)
						return
					}

					if actualResult.Success != tt.then.result.Success {
						t.Errorf("期待されたSuccess値と異なります。expected: %t, actual: %t", tt.then.result.Success, actualResult.Success)
					}

					if actualResult.ResourceID != tt.then.result.ResourceID {
						t.Errorf("期待されたResourceIDと異なります。expected: %s, actual: %s", tt.then.result.ResourceID, actualResult.ResourceID)
					}
				}
			}
		})
	}
}

func TestHotelBookingRequest_Validate(t *testing.T) {
	tests := []struct {
		name  string
		given HotelBookingRequest
		when  string
		then  bool
	}{
		{
			name: "正常なリクエスト",
			given: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "validate",
			then: false, // エラーなし
		},
		{
			name: "BookingIDが空",
			given: HotelBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			when: "validate",
			then: true, // エラーあり
		},
		{
			name: "UserIDが空",
			given: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "",
				HotelID:   "hotel-789",
			},
			when: "validate",
			then: true, // エラーあり
		},
		{
			name: "HotelIDが空",
			given: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "",
			},
			when: "validate",
			then: true, // エラーあり
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.given.Validate()

			// Then
			hasError := err != nil
			if hasError != tt.then {
				t.Errorf("期待されたバリデーション結果と異なります。expected error: %t, actual error: %t", tt.then, hasError)
			}
		})
	}
}
