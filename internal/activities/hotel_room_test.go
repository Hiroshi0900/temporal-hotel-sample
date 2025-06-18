package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テストケースについて
// 正常系:
//   - 正常なリクエストがされた場合、ホテル予約処理が完了する
//   - booking-duplicateの時、冪等性が保証される
//
// 異常系:
//   - BookingIDが空の時、Businessエラーが返却される
//   - UserIDが空の時、Businessエラーが返却される
//   - HotelIDが空の時、Businessエラーが返却される
//   - booking-network-errorの時、Serverエラーが返却される
//   - booking-fullの時、Businessエラーが返却される
func Test_HotelRoomBookingActivity(t *testing.T) {
	testcases := map[string]struct {
		request        HotelBookingRequest
		expectedResult *HotelBookingResult
		expectedErr    error
	}{
		"正常系: 想定通りのリクエストが来た時、ホテル予約が成功する": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedResult: &HotelBookingResult{
				Success:    true,
				ResourceID: "room-123",
				Message:    "ホテルルーム予約が完了しました",
			},
			expectedErr: nil,
		},
		"正常系: booking-duplicateの時、冪等性が保証される": {
			request: HotelBookingRequest{
				BookingID: "booking-duplicate",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedResult: &HotelBookingResult{
				Success:    true,
				ResourceID: "room-duplicate",
				Message:    "既に予約済みです",
			},
			expectedErr: nil,
		},
		"異常系: BookingIDが空の時、Businessエラーが返却される": {
			request: HotelBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "BookingID is required",
				Code:    "INVALID_BOOKING_ID",
			},
		},
		"異常系: UserIDが空の時、Businessエラーが返却される": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "",
				HotelID:   "hotel-789",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "UserID is required",
				Code:    "INVALID_USER_ID",
			},
		},
		"異常系: HotelIDが空の時、Businessエラーが返却される": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "HotelID is required",
				Code:    "INVALID_HOTEL_ID",
			},
		},
		"異常系: booking-network-errorの時、Serverエラーが返却される": {
			request: HotelBookingRequest{
				BookingID: "booking-network-error",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedResult: nil,
			expectedErr: &ServerError{
				Message: "ネットワークエラーが発生しました",
				Code:    "NETWORK_ERROR",
			},
		},
		"異常系: booking-fullの時、Businessエラーが返却される": {
			request: HotelBookingRequest{
				BookingID: "booking-full",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "指定されたホテルは満室です",
				Code:    "HOTEL_FULL",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// given
			ctx := context.Background()
			mockLogger := &MockLogger{}
			sut := NewHotelActivity(mockLogger)

			// when
			actualResult, actualErr := sut.BookHotel(ctx, tc.request)

			// then
			assert.Equal(t, tc.expectedResult, actualResult)
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}

func TestHotelBookingRequest_Validate(t *testing.T) {
	testcases := map[string]struct {
		request     HotelBookingRequest
		expectedErr error
	}{
		"正常系: 正常なリクエスト": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedErr: nil,
		},
		"異常系: BookingIDが空": {
			request: HotelBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				HotelID:   "hotel-789",
			},
			expectedErr: &BusinessError{
				Message: "BookingID is required",
				Code:    "INVALID_BOOKING_ID",
			},
		},
		"異常系: UserIDが空": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "",
				HotelID:   "hotel-789",
			},
			expectedErr: &BusinessError{
				Message: "UserID is required",
				Code:    "INVALID_USER_ID",
			},
		},
		"異常系: HotelIDが空": {
			request: HotelBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				HotelID:   "",
			},
			expectedErr: &BusinessError{
				Message: "HotelID is required",
				Code:    "INVALID_HOTEL_ID",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// given - テストケースで設定済み

			// when
			actualErr := tc.request.Validate()

			// then
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}
