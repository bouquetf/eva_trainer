package apierror

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupGinTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	return c, w
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(ErrCodeNotFound, "resource not found", http.StatusNotFound)

	if err.Code != ErrCodeNotFound {
		t.Errorf("expected Code %q, got %q", ErrCodeNotFound, err.Code)
	}
	if err.Message != "resource not found" {
		t.Errorf("expected Message %q, got %q", "resource not found", err.Message)
	}
	if err.HTTPCode != http.StatusNotFound {
		t.Errorf("expected HTTPCode %d, got %d", http.StatusNotFound, err.HTTPCode)
	}
	if err.Err != nil {
		t.Errorf("expected Err to be nil, got %v", err.Err)
	}
}

func TestNewAPIErrorWithCause(t *testing.T) {
	cause := errors.New("underlying cause")
	err := NewAPIErrorWithCause(ErrCodeBadRequest, "bad request", http.StatusBadRequest, cause)

	if err.Code != ErrCodeBadRequest {
		t.Errorf("expected Code %q, got %q", ErrCodeBadRequest, err.Code)
	}
	if err.Message != "bad request" {
		t.Errorf("expected Message %q, got %q", "bad request", err.Message)
	}
	if err.HTTPCode != http.StatusBadRequest {
		t.Errorf("expected HTTPCode %d, got %d", http.StatusBadRequest, err.HTTPCode)
	}
	if !errors.Is(err, cause) {
		t.Errorf("expected errors.Is(err, cause) to be true")
	}
}

func TestAPIError_Error(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := NewAPIError(ErrCodeInternal, "something broke", http.StatusInternalServerError)
		if err.Error() != "something broke" {
			t.Errorf("expected %q, got %q", "something broke", err.Error())
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("disk full")
		err := NewAPIErrorWithCause(ErrCodeInternal, "something broke", http.StatusInternalServerError, cause)
		expected := "something broke: disk full"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestAPIError_Unwrap(t *testing.T) {
	t.Run("nil when no cause", func(t *testing.T) {
		err := NewAPIError(ErrCodeInternal, "msg", http.StatusInternalServerError)
		if err.Unwrap() != nil {
			t.Errorf("expected nil, got %v", err.Unwrap())
		}
	})

	t.Run("returns cause when set", func(t *testing.T) {
		cause := errors.New("cause")
		err := NewAPIErrorWithCause(ErrCodeInternal, "msg", http.StatusInternalServerError, cause)
		if err.Unwrap() != cause {
			t.Errorf("expected %v, got %v", cause, err.Unwrap())
		}
	})
}

func TestRespondWithError(t *testing.T) {
	t.Run("writes status and json body", func(t *testing.T) {
		c, w := setupGinTest()
		apiErr := NewAPIError(ErrCodeNotFound, "user not found", http.StatusNotFound)
		RespondWithError(c, apiErr)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}
		if body["error"] != string(ErrCodeNotFound) {
			t.Errorf("expected error %q, got %v", ErrCodeNotFound, body["error"])
		}
		if body["message"] != "user not found" {
			t.Errorf("expected message %q, got %v", "user not found", body["message"])
		}
	})

	t.Run("nil apiErr does nothing", func(t *testing.T) {
		c, w := setupGinTest()
		RespondWithError(c, nil)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for nil error, got %d", http.StatusOK, w.Code)
		}
		if body := w.Body.String(); body != "" {
			t.Errorf("expected empty body for nil error, got %q", body)
		}
	})
}

func TestRespondWithSuccess(t *testing.T) {
	c, w := setupGinTest()
	data := map[string]string{"key": "value"}
	RespondWithSuccess(c, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("expected key %q, got %q", "value", body["key"])
	}
}

func TestErrorStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		{name: "internal_error", code: ErrCodeInternal, expected: http.StatusInternalServerError},
		{name: "not_found", code: ErrCodeNotFound, expected: http.StatusNotFound},
		{name: "bad_request", code: ErrCodeBadRequest, expected: http.StatusBadRequest},
		{name: "unauthorized", code: ErrCodeUnauthorized, expected: http.StatusUnauthorized},
		{name: "forbidden", code: ErrCodeForbidden, expected: http.StatusForbidden},
		{name: "conflict", code: ErrCodeConflict, expected: http.StatusConflict},
		{name: "validation_error", code: ErrCodeValidation, expected: http.StatusUnprocessableEntity},
		{name: "discord_api_error", code: ErrCodeDiscordAPI, expected: http.StatusBadGateway},
		{name: "bad_gateway", code: ErrCodeBadGateway, expected: http.StatusBadGateway},
		{name: "rate_limited", code: ErrCodeRateLimited, expected: http.StatusTooManyRequests},
		{name: "unknown", code: "unknown", expected: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorStatusCode(tt.code); got != tt.expected {
				t.Errorf("ErrorStatusCode(%q) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestAPIError_JSON_Serialization(t *testing.T) {
	err := NewAPIErrorWithCause(ErrCodeValidation, "invalid input", http.StatusUnprocessableEntity, errors.New("hidden cause"))
	data, err2 := json.Marshal(err)
	if err2 != nil {
		t.Fatalf("failed to marshal APIError: %v", err2)
	}

	var raw map[string]json.RawMessage
	if err3 := json.Unmarshal(data, &raw); err3 != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err3)
	}

	if _, ok := raw["error"]; !ok {
		t.Error("expected 'error' key in JSON output")
	}
	if _, ok := raw["message"]; !ok {
		t.Error("expected 'message' key in JSON output")
	}

	if _, ok := raw["httpcode"]; ok {
		t.Error("did not expect 'httpcode' key in JSON output (should be tagged json:\"-\")")
	}
	if _, ok := raw["Err"]; ok {
		t.Error("did not expect 'Err' key in JSON output (should be tagged json:\"-\")")
	}
	if _, ok := raw["http_code"]; ok {
		t.Error("did not expect 'http_code' key in JSON output")
	}
	if _, ok := raw["HTTPCode"]; ok {
		t.Error("did not expect 'HTTPCode' key in JSON output")
	}

	var decoded APIError
	if err4 := json.Unmarshal(data, &decoded); err4 != nil {
		t.Fatalf("failed to unmarshal into APIError: %v", err4)
	}
	if decoded.Code != ErrCodeValidation {
		t.Errorf("expected Code %q, got %q", ErrCodeValidation, decoded.Code)
	}
	if decoded.Message != "invalid input" {
		t.Errorf("expected Message %q, got %q", "invalid input", decoded.Message)
	}
}
