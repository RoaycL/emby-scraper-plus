package model

// Response represents the standard MetaTube API JSON envelope.
type Response[T any] struct {
	Data  T          `json:"data,omitempty"`
	Error *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo represents an API error structure.
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
