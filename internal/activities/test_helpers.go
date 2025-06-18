package activities

// MockLogger テスト用のモックロガー
type MockLogger struct{}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	// テストでは何もしない
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// テストでは何もしない
}
