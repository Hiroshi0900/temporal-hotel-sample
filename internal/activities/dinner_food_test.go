package activities

import (
	"strings"
	"testing"

	"go.temporal.io/sdk/testsuite"
)

func TestDinnerBookingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		given   DinnerBookingRequest
		when    string
		then    struct {
			expectError bool
			errorCode   string
		}
	}{
		{
			name:  "正常なリクエスト",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "user1", MenuType: "course"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{false, ""},
		},
		{
			name:  "BookingIDが空文字",
			given: DinnerBookingRequest{BookingID: "", UserID: "user1", MenuType: "course"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_BOOKING_ID"},
		},
		{
			name:  "BookingIDが空白文字のみ",
			given: DinnerBookingRequest{BookingID: "  ", UserID: "user1", MenuType: "course"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_BOOKING_ID"},
		},
		{
			name:  "UserIDが空文字",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "", MenuType: "course"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_USER_ID"},
		},
		{
			name:  "UserIDが空白文字のみ",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "  ", MenuType: "course"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_USER_ID"},
		},
		{
			name:  "MenuTypeが空文字",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "user1", MenuType: ""},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_MENU_TYPE"},
		},
		{
			name:  "MenuTypeが空白文字のみ",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "user1", MenuType: "  "},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_MENU_TYPE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.given.Validate()

			// Then
			if tt.then.expectError {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
					return
				}
				if businessErr, ok := err.(*BusinessError); ok {
					if businessErr.Code != tt.then.errorCode {
						t.Errorf("期待されたエラーコード = %v, 実際 = %v", tt.then.errorCode, businessErr.Code)
					}
				} else {
					t.Errorf("期待されたビジネスエラーではありません: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
				}
			}
		})
	}
}

func TestDinnerFoodBookingActivity(t *testing.T) {
	tests := []struct {
		name  string
		given DinnerBookingRequest
		when  string
		then  struct {
			expectError bool
			errorType   string
			success     bool
			resourceID  string
		}
	}{
		{
			name:  "正常なディナー食材予約",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "user1", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{false, "", true, "food-123"},
		},
		{
			name:  "バリデーションエラー - BookingIDが空",
			given: DinnerBookingRequest{BookingID: "", UserID: "user1", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "バリデーションエラー - UserIDが空",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "バリデーションエラー - MenuTypeが空",
			given: DinnerBookingRequest{BookingID: "booking1", UserID: "user1", MenuType: ""},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "ビジネスエラー - 食材在庫不足",
			given: DinnerBookingRequest{BookingID: "booking-out-of-stock", UserID: "user1", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "一時的エラー - 外部システム障害",
			given: DinnerBookingRequest{BookingID: "booking-system-error", UserID: "user1", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "temporal_error", false, ""},
		},
		{
			name:  "冪等性テスト - 重複リクエスト",
			given: DinnerBookingRequest{BookingID: "booking-duplicate-dinner", UserID: "user1", MenuType: "course"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{false, "", true, "food-duplicate"},
		},
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(DinnerFoodBookingActivity)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - テスト前のキャッシュクリア（冪等性テスト以外）
			if tt.given.BookingID != "booking-duplicate-dinner" {
				clearDinnerCache(tt.given.BookingID)
			} else {
				// 冪等性テストの場合は事前にキャッシュを設定
				prepareDinnerCache(tt.given.BookingID)
			}

			// When
			val, err := env.ExecuteActivity(DinnerFoodBookingActivity, tt.given)

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
				case "business_error":
					if !strings.Contains(errorMsg, "BookingID is required") &&
						!strings.Contains(errorMsg, "UserID is required") &&
						!strings.Contains(errorMsg, "MenuType is required") &&
						!strings.Contains(errorMsg, "指定されたメニューの食材が在庫不足です") {
						t.Errorf("期待されたBusinessErrorではありません: %s", errorMsg)
					}
				case "temporal_error":
					if !strings.Contains(errorMsg, "外部システムで障害が発生しました") {
						t.Errorf("期待されたTemporalErrorではありません: %s", errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var result DinnerBookingResult
				if err := val.Get(&result); err != nil {
					t.Errorf("結果の取得に失敗しました: %v", err)
					return
				}

				if result.Success != tt.then.success {
					t.Errorf("期待された成功フラグ = %v, 実際 = %v", tt.then.success, result.Success)
				}

				if result.ResourceID != tt.then.resourceID {
					t.Errorf("期待されたResourceID = %v, 実際 = %v", tt.then.resourceID, result.ResourceID)
				}
			}
		})
	}
}

// テスト用のキャッシュクリア関数
func clearDinnerCache(bookingID string) {
	delete(dinnerCache, bookingID)
}

// 冪等性テスト用のキャッシュ準備関数
func prepareDinnerCache(bookingID string) {
	dinnerCache[bookingID] = &DinnerBookingResult{
		Success:    true,
		ResourceID: "food-duplicate",
		Message:    "既に予約済みです",
	}
}
