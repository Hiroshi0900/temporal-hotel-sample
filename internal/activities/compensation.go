package activities

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// CompensationResult 補償処理結果
type CompensationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// hotelCompensationCache ホテルルーム補償処理のキャッシュ
var hotelCompensationCache = make(map[string]*CompensationResult)

// dinnerCompensationCache ディナー食材補償処理のキャッシュ  
var dinnerCompensationCache = make(map[string]*CompensationResult)

// parkingCompensationCache 駐車場補償処理のキャッシュ
var parkingCompensationCache = make(map[string]*CompensationResult)

// CompensateHotelRoomActivity ホテルルーム補償アクティビティ
func CompensateHotelRoomActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ホテルルーム補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := hotelCompensationCache[bookingID]; exists {
		logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. ホテル予約システムAPIを呼び出して予約をキャンセル
	// 2. 料金の返金処理
	// 3. 在庫の復旧処理
	
	// シミュレーション: ログ出力のみ
	logger.Info("ホテルルーム予約をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)
	
	result := &CompensationResult{
		Success: true,
		Message: "ホテルルーム予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	hotelCompensationCache[bookingID] = result

	logger.Info("ホテルルーム補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}

// CompensateDinnerFoodActivity ディナー食材補償アクティビティ
func CompensateDinnerFoodActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ディナー食材補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := dinnerCompensationCache[bookingID]; exists {
		logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. 食材仕入れシステムAPIを呼び出して注文をキャンセル
	// 2. 仕入れ代金の返金処理
	// 3. 在庫の調整処理
	
	// シミュレーション: ログ出力のみ
	logger.Info("ディナー食材注文をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)
	
	result := &CompensationResult{
		Success: true,
		Message: "ディナー食材予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	dinnerCompensationCache[bookingID] = result

	logger.Info("ディナー食材補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}

// CompensateParkingActivity 駐車場補償アクティビティ
func CompensateParkingActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("駐車場補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := parkingCompensationCache[bookingID]; exists {
		logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. 駐車場管理システムAPIを呼び出して予約をキャンセル
	// 2. 駐車料金の返金処理
	// 3. 駐車スペースの開放処理
	
	// シミュレーション: ログ出力のみ
	logger.Info("駐車場予約をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)
	
	result := &CompensationResult{
		Success: true,
		Message: "駐車場予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	parkingCompensationCache[bookingID] = result

	logger.Info("駐車場補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}
