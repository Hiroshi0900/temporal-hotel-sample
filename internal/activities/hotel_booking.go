package activities

import (
	"context"
	"strings"
)

// HotelBookingRequest ホテル予約リクエスト
type HotelBookingRequest struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	HotelID   string `json:"hotel_id"`
}

// HotelBookingResult ホテル予約結果
type HotelBookingResult struct {
	Success    bool   `json:"success"`
	ResourceID string `json:"resource_id"`
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code"`
}

type HotelActivity struct {
	logger Logger
}

// Validate リクエストの妥当性チェック
func (hr *HotelBookingRequest) Validate() error {
	if strings.TrimSpace(hr.BookingID) == "" {
		return NewBusinessError("BookingID is required", "INVALID_BOOKING_ID")
	}
	if strings.TrimSpace(hr.UserID) == "" {
		return NewBusinessError("UserID is required", "INVALID_USER_ID")
	}
	if strings.TrimSpace(hr.HotelID) == "" {
		return NewBusinessError("HotelID is required", "INVALID_HOTEL_ID")
	}
	return nil
}

// bookingCache 冪等性のための簡単なインメモリキャッシュ
var bookingCache = make(map[string]*HotelBookingResult)

func NewHotelActivity(logger Logger) *HotelActivity {
	return &HotelActivity{
		logger: logger,
	}
}

// BookHotel ホテルルーム予約アクティビティ
func (a *HotelActivity) BookHotel(ctx context.Context, req HotelBookingRequest) (*HotelBookingResult, error) {
	a.logger.Info("ホテルルーム予約アクティビティを開始", "BookingID", req.BookingID)

	// バリデーション
	if err := req.Validate(); err != nil {
		a.logger.Error("リクエストの妥当性チェックに失敗", "Error", err)
		return nil, err
	}

	// 冪等性チェック（既に処理済みかどうか）
	if cached, exists := bookingCache[req.BookingID]; exists {
		a.logger.Info("既に処理済みの予約リクエスト", "BookingID", req.BookingID)
		return cached, nil
	}

	// 特定のBookingIDに基づくシミュレーション
	switch req.BookingID {
	case "booking-network-error":
		// サーバーエラー（ネットワークエラー）をシミュレート
		err := NewServerError("ネットワークエラーが発生しました", "NETWORK_ERROR")
		a.logger.Error("サーバーエラーが発生", "Error", err)
		return nil, err

	case "booking-full":
		// ビジネスエラー（満室）をシミュレート
		err := NewBusinessError("指定されたホテルは満室です", "HOTEL_FULL")
		a.logger.Error("ビジネスエラーが発生", "Error", err)
		return nil, err

	case "booking-duplicate":
		// 冪等性テストのための特別処理
		result := &HotelBookingResult{
			Success:    true,
			ResourceID: "room-duplicate",
			Message:    "既に予約済みです",
		}
		// キャッシュに保存
		bookingCache[req.BookingID] = result
		a.logger.Info("重複リクエストの処理完了", "BookingID", req.BookingID)
		return result, nil

	default:
		// 正常な予約処理
		result := &HotelBookingResult{
			Success:    true,
			ResourceID: "room-123", // 実際のシステムでは動的に生成
			Message:    "ホテルルーム予約が完了しました",
		}

		// キャッシュに保存（冪等性保証）
		bookingCache[req.BookingID] = result

		a.logger.Info("ホテルルーム予約が完了", "BookingID", req.BookingID, "ResourceID", result.ResourceID)
		return result, nil
	}
}

// HotelRoomBookingActivity ワークフロー用アダプター関数
func HotelRoomBookingActivity(ctx context.Context, req HotelBookingRequest) (*HotelBookingResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewHotelActivity(logger)
	return activity.BookHotel(ctx, req)
}
