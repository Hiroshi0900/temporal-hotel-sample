package workflows

import (
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"temporal-hotel-sample/internal/activities"
)

// BookingRequest ホテル予約Sagaの統合リクエスト
type BookingRequest struct {
	BookingID string         `json:"booking_id"`
	UserID    string         `json:"user_id"`
	Hotel     HotelRequest   `json:"hotel"`
	Dinner    DinnerRequest  `json:"dinner"`
	Parking   ParkingRequest `json:"parking"`
}

// HotelRequest ホテル予約サブリクエスト
type HotelRequest struct {
	HotelID  string    `json:"hotel_id"`
	CheckIn  time.Time `json:"check_in,omitempty"`
	CheckOut time.Time `json:"check_out,omitempty"`
	RoomType string    `json:"room_type,omitempty"`
}

// DinnerRequest ディナー予約サブリクエスト
type DinnerRequest struct {
	MenuType string    `json:"menu_type"`
	DateTime time.Time `json:"date_time,omitempty"`
	Guests   int       `json:"guests,omitempty"`
}

// ParkingRequest 駐車場予約サブリクエスト
type ParkingRequest struct {
	SpaceType string    `json:"space_type"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// BookingResult ホテル予約Sagaの統合結果
type BookingResult struct {
	Success       bool                             `json:"success"`
	BookingID     string                           `json:"booking_id"`
	Message       string                           `json:"message"`
	HotelResult   *activities.HotelBookingResult   `json:"hotel_result,omitempty"`
	DinnerResult  *activities.DinnerBookingResult  `json:"dinner_result,omitempty"`
	ParkingResult *activities.ParkingBookingResult `json:"parking_result,omitempty"`
	Compensations []string                         `json:"compensations,omitempty"` // 実行された補償処理
}

// Validate 統合リクエストのバリデーション
func (r *BookingRequest) Validate() error {
	if strings.TrimSpace(r.BookingID) == "" {
		return fmt.Errorf("BookingID is required")
	}
	if strings.TrimSpace(r.UserID) == "" {
		return fmt.Errorf("UserID is required")
	}
	if strings.TrimSpace(r.Hotel.HotelID) == "" {
		return fmt.Errorf("Hotel.HotelID is required")
	}
	if strings.TrimSpace(r.Dinner.MenuType) == "" {
		return fmt.Errorf("Dinner.MenuType is required")
	}
	if strings.TrimSpace(r.Parking.SpaceType) == "" {
		return fmt.Errorf("Parking.SpaceType is required")
	}
	return nil
}

// HotelBookingSaga ホテル予約Sagaワークフロー
func HotelBookingSaga(ctx workflow.Context, request BookingRequest) (*BookingResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ホテル予約Sagaワークフローを開始", "BookingID", request.BookingID, "UserID", request.UserID)

	// リクエストのバリデーション
	if err := request.Validate(); err != nil {
		logger.Error("リクエストのバリデーションに失敗", "Error", err.Error())
		return &BookingResult{
			Success:   false,
			BookingID: request.BookingID,
			Message:   fmt.Sprintf("バリデーションエラー: %s", err.Error()),
		}, nil // ワークフローとしては正常終了、結果でエラーを表現
	}

	// リトライポリシーの設定
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
		NonRetryableErrorTypes: []string{
			"*activities.BusinessError",
			"*activities.ValidationError",
		},
	}

	// アクティビティオプションの設定
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// 結果の初期化
	result := &BookingResult{
		Success:       false,
		BookingID:     request.BookingID,
		Compensations: []string{},
	}

	// Sagaパターンでの補償処理管理
	var compensations Compensations

	// Step 1: ホテルルーム予約
	logger.Info("ステップ 1: ホテルルーム予約を開始", "HotelID", request.Hotel.HotelID)
	hotelRequest := activities.HotelBookingRequest{
		BookingID: request.BookingID,
		UserID:    request.UserID,
		HotelID:   request.Hotel.HotelID,
	}

	var hotelResult activities.HotelBookingResult
	err := workflow.ExecuteActivity(ctx, activities.HotelRoomBookingActivity, hotelRequest).Get(ctx, &hotelResult)
	if err != nil {
		logger.Error("ホテルルーム予約に失敗", "Error", err.Error())
		result.Message = fmt.Sprintf("ホテルルーム予約に失敗: %s", err.Error())
		return result, nil
	}

	result.HotelResult = &hotelResult
	logger.Info("ステップ 1: ホテルルーム予約が完了", "ResourceID", hotelResult.ResourceID)

	// 補償アクティビティの追加
	compensations.AddCompensation(activities.CompensateHotelRoomActivity)

	// Step 2: ディナー食材予約
	logger.Info("ステップ 2: ディナー食材予約を開始", "MenuType", request.Dinner.MenuType)
	dinnerRequest := activities.DinnerBookingRequest{
		BookingID: request.BookingID,
		UserID:    request.UserID,
		MenuType:  request.Dinner.MenuType,
	}

	var dinnerResult activities.DinnerBookingResult
	err = workflow.ExecuteActivity(ctx, activities.DinnerFoodBookingActivity, dinnerRequest).Get(ctx, &dinnerResult)
	if err != nil {
		logger.Error("ディナー食材予約に失敗", "Error", err.Error())
		result.Message = fmt.Sprintf("ディナー食材予約に失敗: %s", err.Error())
		// 補償処理を実行
		logger.Info("補償処理を開始")
		compensations.Compensate(ctx, false) // 順次実行
		return result, nil
	}

	result.DinnerResult = &dinnerResult
	logger.Info("ステップ 2: ディナー食材予約が完了", "ResourceID", dinnerResult.ResourceID)

	// 補償アクティビティの追加
	compensations.AddCompensation(activities.CompensateDinnerFoodActivity)

	// Step 3: 駐車場予約
	logger.Info("ステップ 3: 駐車場予約を開始", "SpaceType", request.Parking.SpaceType)
	parkingRequest := activities.ParkingBookingRequest{
		BookingID: request.BookingID,
		UserID:    request.UserID,
		SpaceType: request.Parking.SpaceType,
	}

	var parkingResult activities.ParkingBookingResult
	err = workflow.ExecuteActivity(ctx, activities.ParkingBookingActivity, parkingRequest).Get(ctx, &parkingResult)
	if err != nil {
		logger.Error("駐車場予約に失敗", "Error", err.Error())
		result.Message = fmt.Sprintf("駐車場予約に失敗: %s", err.Error())

		// 補償処理を実行
		logger.Info("補償処理を開始")
		compensations.Compensate(ctx, false) // 順次実行
		return result, nil
	}

	result.ParkingResult = &parkingResult
	logger.Info("ステップ 3: 駐車場予約が完了", "ResourceID", parkingResult.ResourceID)

	// 補償アクティビティの追加
	compensations.AddCompensation(activities.CompensateParkingActivity)

	// 全て成功した場合
	result.Success = true
	result.Message = "ホテル予約Sagaが正常に完了しました"
	logger.Info("ホテル予約Sagaワークフローが正常完了", "BookingID", request.BookingID)

	return result, nil
}
