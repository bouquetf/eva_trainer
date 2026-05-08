package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "this-is-a-test-secret-that-is-at-least-32-bytes-long!!"

func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	token, exp, err := svc.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if exp.Before(time.Now()) {
		t.Errorf("expected expiration in the future, got %v", exp)
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	svc := NewJWTService(testSecret, -1*time.Hour, 7*24*time.Hour)

	token, _, err := svc.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	_, err = svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTService_InvalidSignature(t *testing.T) {
	svc1 := NewJWTService("secret-key-1-that-is-long-enough-for-testing!", 24*time.Hour, 7*24*time.Hour)
	svc2 := NewJWTService("secret-key-2-that-is-long-enough-for-testing!!", 24*time.Hour, 7*24*time.Hour)

	token, _, err := svc1.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	_, err = svc2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different key")
	}
}

func TestJWTService_MalformedToken(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	_, err := svc.ValidateAccessToken("not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestJWTService_WrongSigningMethod(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "42",
	})
	badToken, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := svc.ValidateAccessToken(badToken)
	if err == nil {
		t.Fatal("expected error for none signing method")
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	token1, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if len(token1) != 64 {
		t.Errorf("expected token of length 64, got %d", len(token1))
	}

	token2, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if token1 == token2 {
		t.Errorf("expected unique tokens, got identical values")
	}
}

func TestJWTService_ValidateAccessToken_ValidatesSubject(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	token, _, err := svc.GenerateAccessToken(123)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	if claims.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", claims.UserID)
	}
}

func TestJWTService_RefreshExpiry(t *testing.T) {
	svc := NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)

	expiry := svc.RefreshExpiry()
	if expiry != 7*24*time.Hour {
		t.Errorf("expected 7d expiry, got %v", expiry)
	}
}
