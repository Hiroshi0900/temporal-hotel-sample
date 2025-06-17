package activities

import (
	"strings"
	"testing"

	"go.temporal.io/sdk/testsuite"
)

func TestParkingBookingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		given   ParkingBookingRequest
		when    string
		then    struct {
			expectError bool
			errorCode   string
		}
	}{
		{
			name:  "正常なリクエスト",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "user1", SpaceType: "standard"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{false, ""},
		},
		{
			name:  "BookingIDが空文字",
			given: ParkingBookingRequest{BookingID: "", UserID: "user1", SpaceType: "standard"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_BOOKING_ID"},
		},
		{
			name:  "BookingIDが空白文字のみ",
			given: ParkingBookingRequest{BookingID: "  ", UserID: "user1", SpaceType: "standard"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_BOOKING_ID"},
		},
		{
			name:  "UserIDが空文字",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "", SpaceType: "standard"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_USER_ID"},
		},
		{
			name:  "UserIDが空白文字のみ",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "  ", SpaceType: "standard"},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_USER_ID"},
		},
		{
			name:  "SpaceTypeが空文字",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "user1", SpaceType: ""},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_SPACE_TYPE"},
		},
		{
			name:  "SpaceTypeが空白文字のみ",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "user1", SpaceType: "  "},
			when:  "validate",
			then:  struct{ expectError bool; errorCode string }{true, "INVALID_SPACE_TYPE"},
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

func TestParkingBookingActivity(t *testing.T) {
	tests := []struct {
		name  string
		given ParkingBookingRequest
		when  string
		then  struct {
			expectError bool
			errorType   string
			success     bool
			resourceID  string
		}
	}{
		{
			name:  "正常な駐車場予約",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "user1", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{false, "", true, "parking-123"},
		},
		{
			name:  "バリデーションエラー - BookingIDが空",
			given: ParkingBookingRequest{BookingID: "", UserID: "user1", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "バリデーションエラー - UserIDが空",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "バリデーションエラー - SpaceTypeが空",
			given: ParkingBookingRequest{BookingID: "booking1", UserID: "user1", SpaceType: ""},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "ビジネスエラー - 駐車場満車",
			given: ParkingBookingRequest{BookingID: "booking-full", UserID: "user1", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "business_error", false, ""},
		},
		{
			name:  "一時的エラー - 管理システム接続エラー",
			given: ParkingBookingRequest{BookingID: "booking-connection-error", UserID: "user1", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{true, "temporal_error", false, ""},
		},
		{
			name:  "冪等性テスト - 重複リクエスト",
			given: ParkingBookingRequest{BookingID: "booking-duplicate-parking", UserID: "user1", SpaceType: "standard"},
			when:  "execute_activity",
			then:  struct{ expectError bool; errorType string; success bool; resourceID string }{false, "", true, "parking-duplicate"},
		},
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(ParkingBookingActivity)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - テスト前のキャッシュクリア（冪等性テスト以外）
			if tt.given.BookingID != "booking-duplicate-parking" {
				clearParkingCache(tt.given.BookingID)
			} else {
				// 冪等性テストの場合は事前にキャッシュを設定
				prepareParkingCache(tt.given.BookingID)
			}

			// When
			val, err := env.ExecuteActivity(ParkingBookingActivity, tt.given)

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
						!strings.Contains(errorMsg, "SpaceType is required") &&
						!strings.Contains(errorMsg, "指定された駐車場は満車です") {
						t.Errorf("期待されたBusinessErrorではありません: %s", errorMsg)
					}
				case "temporal_error":
					if !strings.Contains(errorMsg, "駐車場管理システムへの接続に失敗しました") {
						t.Errorf("期待されたTemporalErrorではありません: %s", errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラーが発生しました: %v", err)
					return
				}

				var result ParkingBookingResult
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
func clearParkingCache(bookingID string) {
	delete(parkingCache, bookingID)
}

// 冪等性テスト用のキャッシュ準備関数
func prepareParkingCache(bookingID string) {
	parkingCache[bookingID] = &ParkingBookingResult{
		Success:    true,
		ResourceID: "parking-duplicate",
		Message:    "既に予約済みです",
	}
}
