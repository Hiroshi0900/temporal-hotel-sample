# Temporal Hotel Booking System - 設計・実装計画

## 概要
Temporalを使用したホテル予約システムのSagaパターン実装。
3つのリソース（ホテルルーム、ディナー食材、駐車場）を順次確保し、失敗時には補償処理を実行する。

## アプリケーション全体のファイル構成

```
temporal-hotel-sample/
├── cmd/
│   └── server/
│       └── main.go                    # アプリケーションエントリーポイント
├── internal/
│   ├── activities/
│   │   ├── hotel_room.go             # ホテルルーム予約アクティビティ + モデル
│   │   ├── hotel_room_test.go        # ホテルルームアクティビティテスト
│   │   ├── dinner_food.go            # ディナー食材確保アクティビティ + モデル
│   │   ├── dinner_food_test.go       # ディナー食材アクティビティテスト
│   │   ├── parking.go                # 駐車場予約アクティビティ + モデル
│   │   ├── parking_test.go           # 駐車場アクティビティテスト
│   │   ├── compensation.go           # 補償アクティビティ
│   │   └── compensation_test.go      # 補償アクティビティテスト
│   ├── workflows/
│   │   ├── hotel_booking_saga.go     # Sagaワークフロー
│   │   ├── hotel_booking_saga_test.go # Sagaワークフローテスト
│   │   ├── saga.go                   # Compensationsパターン実装
│   │   └── saga_test.go              # Compensationsパターンテスト
│   └── config/
│       ├── temporal.go               # Temporal設定
│       └── temporal_test.go          # Temporal設定テスト
├── Makefile                          # テスト・リント実行用
├── go.mod
├── go.sum
├── docker-compose.yml               # Temporal Server起動用
└── README.md
```

## 設計方針

### 1. Sagaパターンの実装
- **Orchestra型Saga**: 中央のSagaワークフローが各アクティビティを順次実行
- **補償処理**: 提供されたCompensationsパターンを使用
  ```go
  type Compensations []any
  func (s *Compensations) AddCompensation(activity any)
  func (s Compensations) Compensate(ctx workflow.Context, inParallel bool)
  ```
- **順序**: Hotel Room → Dinner Food → Parking の順で実行
- **失敗時**: Compensations.Compensate()で逆順に補償処理を実行

### 2. リトライポリシー
- **指数バックオフ**: 初期間隔1秒、最大間隔60秒
- **最大リトライ回数**: 3回
- **リトライ対象エラー**: 一時的なエラー（ネットワークエラー、タイムアウトなど）
- **非リトライエラー**: ビジネスロジックエラー（重複予約など）

### 3. 冪等性の考慮
- **アクティビティID**: 各アクティビティに一意のIDを付与
- **状態管理**: 既に処理済みの場合は結果を返すのみ
- **補償処理**: 複数回実行されても安全な実装

## テスト設計

### 1. Unit Tests（テーブルケース形式）

#### Activities Tests（各アクティビティファイルと同じディレクトリに配置）

##### hotel_room_test.go
```go
func TestHotelRoomActivity(t *testing.T) {
    tests := []struct {
        name     string
        given    HotelBookingRequest  // アクティビティ内でモデル定義
        when     string // "execute_activity"
        then     struct {
            expectError bool
            errorType   string
        }
    }{
        {
            name: "正常なホテルルーム予約",
            given: HotelBookingRequest{UserID: "user1", HotelID: "hotel1"},
            when: "execute_activity",
            then: struct{expectError bool; errorType string}{false, ""},
        },
        {
            name: "一時的エラー（リトライ対象）",
            given: HotelBookingRequest{UserID: "user_temp_error"},
            when: "execute_activity", 
            then: struct{expectError bool; errorType string}{true, "temporal_error"},
        },
        // 他のテストケース...
    }
}
```

##### dinner_food_test.go / parking_test.go
- 同様の構造でテーブルケース実装（各アクティビティファイルに対応するモデルを使用）

##### compensation_test.go
- 各補償アクティビティの冪等性テスト
- 既に補償済みの場合の処理テスト

#### Workflow Tests（各ワークフローファイルと同じディレクトリに配置）

##### hotel_booking_saga_test.go
```go
func TestHotelBookingSaga(t *testing.T) {
    tests := []struct {
        name     string
        given    struct {
            request BookingRequest
            activityResults map[string]error // アクティビティの結果をモック
        }
        when     string // "execute_workflow"
        then     struct {
            expectSuccess bool
            compensationCalls []string // 呼ばれるべき補償処理
        }
    }{
        {
            name: "全アクティビティ成功",
            given: struct{...}{
                request: BookingRequest{UserID: "user1"},
                activityResults: map[string]error{
                    "hotel_room": nil,
                    "dinner_food": nil, 
                    "parking": nil,
                },
            },
            when: "execute_workflow",
            then: struct{...}{true, []string{}},
        },
        // 他のテストケース...
    }
}
```

##### saga_test.go
- Compensationsパターンのテスト
- 補償処理の順序テスト
- 並列実行オプションのテスト

**テーブルケースを選択する理由**:
- 複数の成功/失敗パターンを網羅的にテストできる
- リトライポリシーの異なる条件を効率的にテスト可能
- Sagaパターンの様々な失敗パターンを整理して検証できる

## 実装に向けた必要なタスク

### Phase 1: 基盤セットアップ
1. **go.mod初期化とTemporal依存関係追加**
2. **Dockerによる開発環境構築**
   - docker-compose.ymlでTemporal Server起動
3. **基本ディレクトリ構造作成**
4. **Makefile作成**（test、lint実行コマンド）

### Phase 2: 基盤実装（TDD）
1. **config/temporal.go** + テスト - Temporal設定とリトライポリシー
2. **workflows/saga.go** + テスト - Compensationsパターン実装

### Phase 3: アクティビティ実装（TDD）
1. **activities/hotel_room.go** + テスト
   - HotelBookingRequest/Resultモデル
   - HotelRoomBookingActivity
2. **activities/dinner_food.go** + テスト
   - DinnerBookingRequest/Resultモデル
   - DinnerFoodBookingActivity  
3. **activities/parking.go** + テスト
   - ParkingBookingRequest/Resultモデル
   - ParkingBookingActivity
4. **activities/compensation.go** + テスト
   - CompensateHotelRoomActivity
   - CompensateDinnerFoodActivity
   - CompensateParkingActivity

### Phase 4: ワークフロー実装（TDD）
1. **workflows/hotel_booking_saga.go** + テスト
   - 統合的な予約リクエストモデル
   - リトライポリシー設定
   - 順次実行とエラーハンドリング
   - 補償処理の実装

### Phase 5: エントリーポイント
1. **cmd/server/main.go** - アプリケーション起動
2. **最終テスト実行とリント確認**

## 技術的な考慮事項

### リトライポリシー設定
```go
retryPolicy := &temporal.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2.0,
    MaximumInterval:    time.Minute,
    MaximumAttempts:    3,
    NonRetryableErrorTypes: []string{
        "BusinessLogicError",
        "InvalidRequestError",
    },
}
```

### 冪等性実装
- アクティビティIDによる重複実行防止
- 状態チェックによる既処理判定
- 補償処理の安全な実行

### エラーハンドリング
- 一時的エラーと永続的エラーの区別
- 適切なエラータイプの定義
- ログによる処理状況の可視化

## 進捗状況

### Phase 1: 基盤セットアップ
- [ ] go.mod初期化とTemporal依存関係追加
- [ ] Dockerによる開発環境構築（docker-compose.yml）
- [ ] 基本ディレクトリ構造作成
- [ ] Makefile作成（test、lint実行コマンド）

### Phase 2: 基盤実装（TDD）
- [x] config/temporal.go + テスト - Temporal設定とリトライポリシー
- [x] workflows/saga.go + テスト - Compensationsパターン実装

### Phase 3: アクティビティ実装（TDD）
- [x] activities/hotel_room.go + テスト（モデル含む）
- [x] activities/dinner_food.go + テスト（モデル含む）
- [x] activities/parking.go + テスト（モデル含む）
- [x] activities/compensation.go + テスト（実装完了）

### Phase 4: ワークフロー実装（TDD）
- [x] workflows/hotel_booking_saga.go + テスト（実装完了）
  - **状況**: Sagaワークフローの本体とテストが完全に実装完了
  - **解決済み**: 
    - モックの引数マッチングエラーを解決
    - Temporalテストフレームワークに適したモック設定を実装
    - 補償処理のテスト設計をsaga.goの設計と整合させて完了
    - ActivityOptionsでのタイムアウト設定不足を修正

### Phase 5: エントリーポイント
- [x] cmd/server/main.go - アプリケーション起動
- [x] 最終テスト実行とリント確認

## 実装メモ
<!-- ここに実装中の気づきや課題を記録 -->
- Phase 2完了: Sagaパターンのテストに異常系を追加済み
- 型安全性改善: anyからinterface{}に変更
- Timeoutエラーは実際のアクティビティ実行時にActivityOptionsで解決予定
- Phase 3-1完了: HotelRoomアクティビティでTDDパターン確立
- Temporalテスト環境のエラーハンドリング: カスタムエラーがActivityErrorにラップされることを確認

## テスト実行結果
<!-- 各フェーズでのテスト結果を記録 -->
- Phase 2: workflows/saga_test.go - 全テストPASS（6テスト）
  - 正常系: 順次・並列実行
  - 異常系: エラー発生時の処理継続
  - エッジケース: 空の補償処理
- Phase 3-1: activities/hotel_room_test.go - 全テストPASS（9テスト）
  - 正常系: ホテルルーム予約成功
  - 異常系: バリデーションエラー、ビジネスエラー、一時的エラー
  - 冪等性: 重複リクエストの処理
  - バリデーション: 各フィールドの妥当性チェック
- Phase 3-2: activities/dinner_food_test.go - 全テストPASS（7テスト）
  - 正常系: ディナー食材予約成功
  - 異常系: バリデーションエラー（BookingID、UserID、MenuType）、ビジネスエラー、一時的エラー
  - 冪等性: 重複リクエストの処理
  - エラーハンドリング: Temporalテスト環境でのActivityErrorラップ対応
- Phase 3-3: activities/parking_test.go - 全テストPASS（7テスト）
  - 正常系: 駐車場予約成功
  - 異常系: バリデーションエラー（BookingID、UserID、SpaceType）、ビジネスエラー、一時的エラー
  - 冪等性: 重複リクエストの処理
  - 同様のエラーハンドリングパターンを適用
