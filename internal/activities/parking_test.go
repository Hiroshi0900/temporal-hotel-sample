package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テストケースについて
// 正常系:
//   - 正常なリクエストがされた場合、駐車場予約処理が完了する
//   - booking-duplicate-parkingの時、冪等性が保証される
//
// 異常系:
//   - BookingIDが空の時、Businessエラーが返却される
//   - UserIDが空の時、Businessエラーが返却される
//   - SpaceTypeが空の時、Businessエラーが返却される
//   - booking-connection-errorの時、Serverエラーが返却される
//   - booking-fullの時、Businessエラーが返却される
func Test_ParkingBookingActivity(t *testing.T) {
	testcases := map[string]struct {
		request        ParkingBookingRequest
		expectedResult *ParkingBookingResult
		expectedErr    error
	}{
		"正常系: 想定通りのリクエストが来た時、駐車場予約が成功する": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedResult: &ParkingBookingResult{
				Success:    true,
				ResourceID: "parking-123",
				Message:    "駐車場予約が完了しました",
			},
			expectedErr: nil,
		},
		"正常系: booking-duplicate-parkingの時、冪等性が保証される": {
			request: ParkingBookingRequest{
				BookingID: "booking-duplicate-parking",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedResult: &ParkingBookingResult{
				Success:    true,
				ResourceID: "parking-duplicate",
				Message:    "既に予約済みです",
			},
			expectedErr: nil,
		},
		"異常系: BookingIDが空の時、Businessエラーが返却される": {
			request: ParkingBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "BookingID is required",
				Code:    "INVALID_BOOKING_ID",
			},
		},
		"異常系: UserIDが空の時、Businessエラーが返却される": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "",
				SpaceType: "standard",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "UserID is required",
				Code:    "INVALID_USER_ID",
			},
		},
		"異常系: SpaceTypeが空の時、Businessエラーが返却される": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				SpaceType: "",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "SpaceType is required",
				Code:    "INVALID_SPACE_TYPE",
			},
		},
		"異常系: booking-connection-errorの時、Serverエラーが返却される": {
			request: ParkingBookingRequest{
				BookingID: "booking-connection-error",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedResult: nil,
			expectedErr: &ServerError{
				Message: "駐車場管理システムへの接続に失敗しました",
				Code:    "CONNECTION_ERROR",
			},
		},
		"異常系: booking-fullの時、Businessエラーが返却される": {
			request: ParkingBookingRequest{
				BookingID: "booking-full",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedResult: nil,
			expectedErr: &BusinessError{
				Message: "指定された駐車場は満車です",
				Code:    "PARKING_FULL",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			// given
			ctx := context.Background()
			mockLogger := &MockLogger{}
			sut := NewParkingActivity(mockLogger)

			// when
			actualResult, actualErr := sut.BookParking(ctx, tc.request)

			// then
			assert.Equal(t, tc.expectedResult, actualResult)
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}

func TestParkingBookingRequest_Validate(t *testing.T) {
	testcases := map[string]struct {
		request     ParkingBookingRequest
		expectedErr error
	}{
		"正常系: 正常なリクエスト": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedErr: nil,
		},
		"異常系: BookingIDが空": {
			request: ParkingBookingRequest{
				BookingID: "",
				UserID:    "user-456",
				SpaceType: "standard",
			},
			expectedErr: &BusinessError{
				Message: "BookingID is required",
				Code:    "INVALID_BOOKING_ID",
			},
		},
		"異常系: UserIDが空": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "",
				SpaceType: "standard",
			},
			expectedErr: &BusinessError{
				Message: "UserID is required",
				Code:    "INVALID_USER_ID",
			},
		},
		"異常系: SpaceTypeが空": {
			request: ParkingBookingRequest{
				BookingID: "booking-123",
				UserID:    "user-456",
				SpaceType: "",
			},
			expectedErr: &BusinessError{
				Message: "SpaceType is required",
				Code:    "INVALID_SPACE_TYPE",
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
