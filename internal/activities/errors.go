package activities

// BusinessError ビジネスロジックエラー（リトライ不可）
type BusinessError struct {
	Message string
	Code    string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// NewBusinessError ビジネスエラーを作成
func NewBusinessError(message, code string) *BusinessError {
	return &BusinessError{
		Message: message,
		Code:    code,
	}
}

// ServerError サーバーエラー（リトライ可能）
type ServerError struct {
	Message string
	Code    string
}

func (e *ServerError) Error() string {
	return e.Message
}

// NewServerError サーバーエラーを作成
func NewServerError(message, code string) *ServerError {
	return &ServerError{
		Message: message,
		Code:    code,
	}
}

// UnknownError 分類不可エラー（リトライ可能）
type UnknownError struct {
	Message string
	Code    string
}

func (e *UnknownError) Error() string {
	return e.Message
}

// NewUnknownError 分類不可エラーを作成
func NewUnknownError(message, code string) *UnknownError {
	return &UnknownError{
		Message: message,
		Code:    code,
	}
}
