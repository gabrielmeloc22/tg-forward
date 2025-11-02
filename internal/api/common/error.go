package common

type ApiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Meta    any    `json:"meta,omitempty"`
}
