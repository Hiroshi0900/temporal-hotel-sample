package activities

import "context"

// parkingCompensationCache 駐車場補償処理のキャッシュ
var parkingCompensationCache = make(map[string]*CompensationResult)

// CompensateParkingActivity 駐車場補償アクティビティ
func (a *ParkingActivity) CompensateParking(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	a.logger.Info("駐車場補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := parkingCompensationCache[bookingID]; exists {
		a.logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. 駐車場管理システムAPIを呼び出して予約をキャンセル
	// 2. 駐車料金の返金処理
	// 3. 駐車スペースの開放処理
	
	// シミュレーション: ログ出力のみ
	a.logger.Info("駐車場予約をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)
	
	result := &CompensationResult{
		Success: true,
		Message: "駐車場予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	parkingCompensationCache[bookingID] = result

	a.logger.Info("駐車場補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}

// CompensateParkingActivity ワークフロー用アダプター関数
func CompensateParkingActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewParkingActivity(logger)
	return activity.CompensateParking(ctx, bookingID, resourceID)
}