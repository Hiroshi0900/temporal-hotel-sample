package activities

import (
	"context"
	"strings"
)

type (
	// DinnerBookingRequest ディナー食材予約リクエスト
	DinnerBookingRequest struct {
		BookingID string `json:"booking_id"`
		UserID    string `json:"user_id"`
		MenuType  string `json:"menu_type"`
	}
	DinnerBookingResult struct {
		Success    bool   `json:"success"`
		ResourceID string `json:"resource_id"`
		Message    string `json:"message"`
		ErrorCode  string `json:"error_code"`
	}

	DinnerActivity struct {
		logger Logger
	}
)

// Validate リクエストの妥当性チェック
func (dr *DinnerBookingRequest) Validate() error {
	if strings.TrimSpace(dr.BookingID) == "" {
		return NewBusinessError("BookingID is required", "INVALID_BOOKING_ID")
	}
	if strings.TrimSpace(dr.UserID) == "" {
		return NewBusinessError("UserID is required", "INVALID_USER_ID")
	}
	if strings.TrimSpace(dr.MenuType) == "" {
		return NewBusinessError("MenuType is required", "INVALID_MENU_TYPE")
	}
	return nil
}

// dinnerCache 冪等性のための簡単なインメモリキャッシュ
var dinnerCache = make(map[string]*DinnerBookingResult)

func NewDinnerActivity(logger Logger) *DinnerActivity {
	return &DinnerActivity{
		logger: logger,
	}
}

// DinnerFoodBookingActivity ワークフロー用アダプター関数
func DinnerFoodBookingActivity(ctx context.Context, req DinnerBookingRequest) (*DinnerBookingResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewDinnerActivity(logger)
	return activity.BookDinner(ctx, req)
}

func (a *DinnerActivity) BookDinner(ctx context.Context, req DinnerBookingRequest) (*DinnerBookingResult, error) {
	a.logger.Info("ディナー食材予約アクティビティを開始", "BookingID", req.BookingID)

	// バリデーション
	if err := req.Validate(); err != nil {
		a.logger.Error("リクエストの妥当性チェックに失敗", "Error", err)
		return nil, err
	}

	// 冪等性チェック（既に処理済みかどうか）
	if cached, exists := dinnerCache[req.BookingID]; exists {
		a.logger.Info("既に処理済みの予約リクエスト", "BookingID", req.BookingID)
		return cached, nil
	}

	// 特定のBookingIDに基づくシミュレーション
	switch req.BookingID {
	case "booking-system-error":
		// サーバーエラー（外部システム障害）をシミュレート
		err := NewServerError("外部システムで障害が発生しました", "SYSTEM_ERROR")
		a.logger.Error("サーバーエラーが発生", "Error", err)
		return nil, err

	case "booking-out-of-stock":
		// ビジネスエラー（食材在庫不足）をシミュレート
		err := NewBusinessError("指定されたメニューの食材が在庫不足です", "OUT_OF_STOCK")
		a.logger.Error("ビジネスエラーが発生", "Error", err)
		return nil, err

	case "booking-duplicate-dinner":
		// 冪等性テストのための特別処理
		result := &DinnerBookingResult{
			Success:    true,
			ResourceID: "food-duplicate",
			Message:    "既に予約済みです",
		}
		// キャッシュに保存
		dinnerCache[req.BookingID] = result
		a.logger.Info("重複リクエストの処理完了", "BookingID", req.BookingID)
		return result, nil

	default:
		// 正常な予約処理
		result := &DinnerBookingResult{
			Success:    true,
			ResourceID: "food-123", // 実際のシステムでは動的に生成
			Message:    "ディナー食材予約が完了しました",
		}

		// キャッシュに保存（冪等性保証）
		dinnerCache[req.BookingID] = result

		a.logger.Info("ディナー食材予約が完了", "BookingID", req.BookingID, "ResourceID", result.ResourceID)
		return result, nil
	}
}
