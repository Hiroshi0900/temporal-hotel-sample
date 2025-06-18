package activities

import "context"

// hotelCompensationCache ホテルルーム補償処理のキャッシュ
var hotelCompensationCache = make(map[string]*CompensationResult)

// CompensateHotelRoomActivity ホテルルーム補償アクティビティ
func (a *HotelActivity) CompensateHotel(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	a.logger.Info("ホテルルーム補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := hotelCompensationCache[bookingID]; exists {
		a.logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. ホテル予約システムAPIを呼び出して予約をキャンセル
	// 2. 料金の返金処理
	// 3. 在庫の復旧処理
	
	// シミュレーション: ログ出力のみ
	a.logger.Info("ホテルルーム予約をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)
	
	result := &CompensationResult{
		Success: true,
		Message: "ホテルルーム予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	hotelCompensationCache[bookingID] = result

	a.logger.Info("ホテルルーム補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}

// CompensateHotelRoomActivity ワークフロー用アダプター関数
func CompensateHotelRoomActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewHotelActivity(logger)
	return activity.CompensateHotel(ctx, bookingID, resourceID)
}