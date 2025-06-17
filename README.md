# Temporal Hotel Booking System

Temporalを使用したホテル予約システムのSagaパターン実装サンプル。

## 概要

3つのリソース（ホテルルーム、ディナー食材、駐車場）を順次確保し、失敗時には補償処理を実行するSagaパターンを実装。

## セットアップ

### 前提条件
- Go 1.24以上
- Docker & Docker Compose
- golangci-lint

### 開発環境起動

1. Temporal Serverの起動
```bash
docker-compose up -d
```

2. 依存関係のインストール
```bash
make deps
make install-tools
```

3. アプリケーションの実行
```bash
make run
```

## 開発

### テスト実行
```bash
make test
```

### リント実行
```bash
make lint
```

### すべてのチェック
```bash
make check
```

## アーキテクチャ

- **Sagaパターン**: 分散トランザクションの実装
- **リトライポリシー**: 一時的な障害に対する自動リトライ
- **補償処理**: 失敗時の状態復旧
- **冪等性**: 重複実行に対する安全性

## アクセス

- Temporal Web UI: http://localhost:8080
- PostgreSQL: localhost:5432
