package activities

import "context"

// dinnerCompensationCache ディナー食材補償処理のキャッシュ
var dinnerCompensationCache = make(map[string]*CompensationResult)

// CompensateDinnerFoodActivity ワークフロー用アダプター関数
func CompensateDinnerFoodActivity(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	logger := NewTemporalLogger(ctx)
	activity := NewDinnerActivity(logger)
	return activity.CompensateDinner(ctx, bookingID, resourceID)
}

func (a *DinnerActivity) CompensateDinner(ctx context.Context, bookingID string, resourceID string) (*CompensationResult, error) {
	a.logger.Info("ディナー食材補償処理を開始", "BookingID", bookingID, "ResourceID", resourceID)

	// 冪等性チェック（既に補償済みかどうか）
	if cached, exists := dinnerCompensationCache[bookingID]; exists {
		a.logger.Info("既に補償済みの予約", "BookingID", bookingID)
		return cached, nil
	}

	// 実際のシステムでは以下のような処理を行う：
	// 1. 食材仕入れシステムAPIを呼び出して注文をキャンセル
	// 2. 仕入れ代金の返金処理
	// 3. 在庫の調整処理

	// シミュレーション: ログ出力のみ
	a.logger.Info("ディナー食材注文をキャンセルしました", "BookingID", bookingID, "ResourceID", resourceID)

	result := &CompensationResult{
		Success: true,
		Message: "ディナー食材予約の補償処理が完了しました",
	}

	// キャッシュに保存（冪等性保証）
	dinnerCompensationCache[bookingID] = result

	a.logger.Info("ディナー食材補償処理が完了", "BookingID", bookingID, "ResourceID", resourceID)
	return result, nil
}
