package activities

import (
	"testing"

	"go.temporal.io/sdk/testsuite"
)

func TestCompensateHotelRoomActivity(t *testing.T) {
	tests := []struct {
		name  string
		given struct {
			bookingID  string
			resourceID string
		}
		when string
		then struct {
			expectError bool
			success     bool
		}
	}{
		{
			name: "正常なホテルルーム補償",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "room-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
		{
			name: "補償対象リソースが見つからない場合",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "room-notfound"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true}, // 冪等性のため成功扱い
		},
		{
			name: "冪等性テスト - 既に補償済み",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking-compensated-hotel", "room-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CompensateHotelRoomActivity)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - 冪等性テストの場合は事前に補償状態を設定
			if tt.given.bookingID == "booking-compensated-hotel" {
				prepareHotelCompensationCache(tt.given.bookingID)
			} else {
				clearHotelCompensationCache(tt.given.bookingID)
			}

			// When
			val, err := env.ExecuteActivity(CompensateHotelRoomActivity, tt.given.bookingID, tt.given.resourceID)

			// Then
			if tt.then.expectError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
					return
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var result CompensationResult
				if err := val.Get(&result); err != nil {
					t.Errorf("結果の取得に失敗しました: %v", err)
					return
				}

				if result.Success != tt.then.success {
					t.Errorf("期待された成功フラグ = %v, 実際 = %v", tt.then.success, result.Success)
				}
			}
		})
	}
}

func TestCompensateDinnerFoodActivity(t *testing.T) {
	tests := []struct {
		name  string
		given struct {
			bookingID  string
			resourceID string
		}
		when string
		then struct {
			expectError bool
			success     bool
		}
	}{
		{
			name: "正常なディナー食材補償",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "food-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
		{
			name: "補償対象リソースが見つからない場合",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "food-notfound"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true}, // 冪等性のため成功扱い
		},
		{
			name: "冪等性テスト - 既に補償済み",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking-compensated-dinner", "food-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CompensateDinnerFoodActivity)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - 冪等性テストの場合は事前に補償状態を設定
			if tt.given.bookingID == "booking-compensated-dinner" {
				prepareDinnerCompensationCache(tt.given.bookingID)
			} else {
				clearDinnerCompensationCache(tt.given.bookingID)
			}

			// When
			val, err := env.ExecuteActivity(CompensateDinnerFoodActivity, tt.given.bookingID, tt.given.resourceID)

			// Then
			if tt.then.expectError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
					return
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var result CompensationResult
				if err := val.Get(&result); err != nil {
					t.Errorf("結果の取得に失敗しました: %v", err)
					return
				}

				if result.Success != tt.then.success {
					t.Errorf("期待された成功フラグ = %v, 実際 = %v", tt.then.success, result.Success)
				}
			}
		})
	}
}

func TestCompensateParkingActivity(t *testing.T) {
	tests := []struct {
		name  string
		given struct {
			bookingID  string
			resourceID string
		}
		when string
		then struct {
			expectError bool
			success     bool
		}
	}{
		{
			name: "正常な駐車場補償",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "parking-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
		{
			name: "補償対象リソースが見つからない場合",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking1", "parking-notfound"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true}, // 冪等性のため成功扱い
		},
		{
			name: "冪等性テスト - 既に補償済み",
			given: struct {
				bookingID  string
				resourceID string
			}{"booking-compensated-parking", "parking-123"},
			when: "execute_activity",
			then: struct{ expectError bool; success bool }{false, true},
		},
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(CompensateParkingActivity)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - 冪等性テストの場合は事前に補償状態を設定
			if tt.given.bookingID == "booking-compensated-parking" {
				prepareParkingCompensationCache(tt.given.bookingID)
			} else {
				clearParkingCompensationCache(tt.given.bookingID)
			}

			// When
			val, err := env.ExecuteActivity(CompensateParkingActivity, tt.given.bookingID, tt.given.resourceID)

			// Then
			if tt.then.expectError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
					return
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var result CompensationResult
				if err := val.Get(&result); err != nil {
					t.Errorf("結果の取得に失敗しました: %v", err)
					return
				}

				if result.Success != tt.then.success {
					t.Errorf("期待された成功フラグ = %v, 実際 = %v", tt.then.success, result.Success)
				}
			}
		})
	}
}

// テスト用のキャッシュクリア・準備関数
func clearHotelCompensationCache(bookingID string) {
	delete(hotelCompensationCache, bookingID)
}

func prepareHotelCompensationCache(bookingID string) {
	hotelCompensationCache[bookingID] = &CompensationResult{
		Success: true,
		Message: "既に補償済みです",
	}
}

func clearDinnerCompensationCache(bookingID string) {
	delete(dinnerCompensationCache, bookingID)
}

func prepareDinnerCompensationCache(bookingID string) {
	dinnerCompensationCache[bookingID] = &CompensationResult{
		Success: true,
		Message: "既に補償済みです",
	}
}

func clearParkingCompensationCache(bookingID string) {
	delete(parkingCompensationCache, bookingID)
}

func prepareParkingCompensationCache(bookingID string) {
	parkingCompensationCache[bookingID] = &CompensationResult{
		Success: true,
		Message: "既に補償済みです",
	}
}
