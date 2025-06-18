package activities

import (
	"context"
	"strings"
)

// ParkingBookingRequest 駐車場予約リクエスト
type ParkingBookingRequest struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	SpaceType string `json:"space_type"`
}

// ParkingBookingResult 駐車場予約結果
type ParkingBookingResult struct {
	Success    bool   `json:"success"`
	ResourceID string `json:"resource_id"`
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code"`
}

type ParkingActivity struct {
	logger Logger
}

// Validate リクエストの妥当性チェック
func (pr *ParkingBookingRequest) Validate() error {
	if strings.TrimSpace(pr.BookingID) == "" {
		return NewBusinessError("BookingID is required", "INVALID_BOOKING_ID")
	}
	if strings.TrimSpace(pr.UserID) == "" {
		return NewBusinessError("UserID is required", "INVALID_USER_ID")
	}
	if strings.TrimSpace(pr.SpaceType) == "" {
		return NewBusinessError("SpaceType is required", "INVALID_SPACE_TYPE")
	}
	return nil
}

// parkingCache 冪等性のための簡単なインメモリキャッシュ
var parkingCache = make(map[string]*ParkingBookingResult)

func NewParkingActivity(logger Logger) *ParkingActivity {
	return &ParkingActivity{
		logger: logger,
	}
}

// BookParking 駐車場予約アクティビティ
func (a *ParkingActivity) BookParking(ctx context.Context, req ParkingBookingRequest) (*ParkingBookingResult, error) {
	a.logger.Info("駐車場予約アクティビティを開始", "BookingID", req.BookingID)

	// バリデーション
	if err := req.Validate(); err != nil {
		a.logger.Error("リクエストの妥当性チェックに失敗", "Error", err)
		return nil, err
	}

	// 冪等性チェック（既に処理済みかどうか）
	if cached, exists := parkingCache[req.BookingID]; exists {
		a.logger.Info("既に処理済みの予約リクエスト", "BookingID", req.BookingID)
		return cached, nil
	}

	// 特定のBookingIDに基づくシミュレーション
	switch req.BookingID {
	case "booking-connection-error":
		// サーバーエラー（駐車場管理システム接続エラー）をシミュレート
		err := NewServerError("駐車場管理システムへの接続に失敗しました", "CONNECTION_ERROR")
		a.logger.Error("サーバーエラーが発生", "Error", err)
		return nil, err

	case "booking-full":
		// ビジネスエラー（駐車場満車）をシミュレート
		err := NewBusinessError("指定された駐車場は満車です", "PARKING_FULL")
		a.logger.Error("ビジネスエラーが発生", "Error", err)
		return nil, err

	case "booking-duplicate-parking":
		// 冪等性テストのための特別処理
		result := &ParkingBookingResult{
			Success:    true,
			ResourceID: "parking-duplicate",
			Message:    "既に予約済みです",
		}
		// キャッシュに保存
		parkingCache[req.BookingID] = result
		a.logger.Info("重複リクエストの処理完了", "BookingID", req.BookingID)
		return result, nil

	default:
		// 正常な予約処理
		result := &ParkingBookingResult{
			Success:    true,
			ResourceID: "parking-123", // 実際のシステムでは動的に生成
			Message:    "駐車場予約が完了しました",
		}
		
		// キャッシュに保存（冪等性保証）
		parkingCache[req.BookingID] = result
		
		a.logger.Info("駐車場予約が完了", "BookingID", req.BookingID, "ResourceID", result.ResourceID)
		return result, nil
	}
}

// ParkingBookingActivity ワークフロー用アダプター関数
func ParkingBookingActivity(ctx context.Context, req ParkingBookingRequest) (*ParkingBookingResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewParkingActivity(logger)
	return activity.BookParking(ctx, req)
}