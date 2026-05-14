package apierror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrCodeInternal     ErrorCode = "internal_error"
	ErrCodeNotFound     ErrorCode = "not_found"
	ErrCodeBadRequest   ErrorCode = "bad_request"
	ErrCodeUnauthorized ErrorCode = "unauthorized"
	ErrCodeForbidden    ErrorCode = "forbidden"
	ErrCodeConflict     ErrorCode = "conflict"
	ErrCodeValidation   ErrorCode = "validation_error"
	ErrCodeDiscordAPI   ErrorCode = "discord_api_error"
	ErrCodeBadGateway   ErrorCode = "bad_gateway"
	ErrCodeRateLimited  ErrorCode = "rate_limited"
)

type APIError struct {
	Code     ErrorCode      `json:"error"`
	Message  string         `json:"message"`
	HTTPCode int            `json:"-"`
	Err      error          `json:"-"`
	Extra    map[string]any `json:"-"`
}

func (e *APIError) WithExtra(key string, value any) *APIError {
	if e.Extra == nil {
		e.Extra = make(map[string]any)
	}
	e.Extra[key] = value
	return e
}

func NewAPIError(code ErrorCode, message string, httpCode int) *APIError {
	return &APIError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
	}
}

func NewAPIErrorWithCause(code ErrorCode, message string, httpCode int, cause error) *APIError {
	return &APIError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
		Err:      cause,
	}
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func RespondWithError(c *gin.Context, apiErr *APIError) {
	if apiErr == nil {
		return
	}
	resp := gin.H{
		"error":   apiErr.Code,
		"message": apiErr.Message,
	}
	for k, v := range apiErr.Extra {
		resp[k] = v
	}
	c.JSON(apiErr.HTTPCode, resp)
	c.Abort()
}

func RespondWithSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func ErrorStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeValidation:
		return http.StatusUnprocessableEntity
	case ErrCodeDiscordAPI, ErrCodeBadGateway:
		return http.StatusBadGateway
	case ErrCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
