package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テストケースについて
// 正常系:
//   - 正常なリクエストがされた場合、予約処理が完了する
//
// 異常系:
//   - BookingIDが空の時、Businessエラーが返却される
//   - BookingIDがbooking-system-error（サーバエラー）の時、Serverエラーが返却される
//   - BookingIDがbooking-out-of-stockの時、Businessエラーが返却される
//   - BookingIDがbooking-duplicateの時、冪等性が保証される
func Test_DinnerFoodBookingActivity(t *testing.T) {
	testcases := map[string]struct {
		request        DinnerBookingRequest
		expectedResult *DinnerBookingResult
		expectedErr    error
	}{
		"正常系: 想定通りのリクエストが来た時、予約が成功する": {
			request: DinnerBookingRequest{
				BookingID: "booking1",
				UserID:    "user1",
				MenuType:  "course",
			},
			expectedResult: &DinnerBookingResult{
				Success:    true,
				ResourceID: "food-123",
				Message:    "ディナー食材予約が完了しました",
			},
			expectedErr: nil,
		},
		"異常系: BookingIDが空の時、Businessエラーが返却される": {
			request: DinnerBookingRequest{
				BookingID: "",
				UserID:    "user1",
				MenuType:  "course",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "BookingID is required",
				Code:    "INVALID_BOOKING_ID",
			},
		},
		"異常系: booking-system-errorの時、Serverエラーが返却される": {
			request: DinnerBookingRequest{
				BookingID: "booking-system-error",
				UserID:    "user1",
				MenuType:  "course",
			},
			expectedResult: nil,
			expectedErr: &ServerError{
				Message: "外部システムで障害が発生しました",
				Code:    "SYSTEM_ERROR",
			},
		},
		"異常系: booking-out-of-stockの時、Businessエラーが返却される": {
			request: DinnerBookingRequest{
				BookingID: "booking-out-of-stock",
				UserID:    "user1",
				MenuType:  "course",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "指定されたメニューの食材が在庫不足です",
				Code:    "OUT_OF_STOCK",
			},
		},
		"異常系: booking-duplicateの時、冪等性が保証される": {
			request: DinnerBookingRequest{
				BookingID: "booking-duplicate-dinner",
				UserID:    "user1",
				MenuType:  "course",
			},
			expectedResult: &DinnerBookingResult{
				Success:    true,
				ResourceID: "food-duplicate",
				Message:    "既に予約済みです",
			},
			expectedErr: nil,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// given
			ctx := context.Background()
			mockLogger := &MockLogger{}
			sut := NewDinnerActivity(mockLogger)

			// when
			actualResult, actualErr := sut.BookDinner(ctx, tc.request)

			// then
			assert.Equal(t, tc.expectedResult, actualResult)
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}
