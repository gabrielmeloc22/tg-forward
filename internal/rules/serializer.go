package rules

type Error struct {
	StatusCode int
	Code       string
	Message    string
	Meta       map[string]any
}

func (e *Error) Error() string {
	return e.Message
}

type ApiErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func NewError(statusCode int, code, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Meta:       nil,
	}
}

func NewErrorWithMeta(statusCode int, code, message string, meta map[string]any) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Meta:       meta,
	}
}
