package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const discordAuthURL = "https://discord.com/api/oauth2/authorize"
const discordTokenURL = "https://discord.com/api/oauth2/token"
const discordUserURL = "https://discord.com/api/users/@me"

const stateExpiry = 5 * time.Minute

type StateClaims struct {
	Redirect string `json:"redirect"`
	jwt.RegisteredClaims
}

type DiscordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
}

func DiscordOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  discordAuthURL,
			TokenURL: discordTokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"identify"},
	}
}

func GenerateOAuthState(jwtSecret []byte, redirectPath string) (string, error) {
	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	claims := StateClaims{
		Redirect: redirectPath,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   hex.EncodeToString(nonce),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(stateExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign state: %w", err)
	}

	return signed, nil
}

func ValidateOAuthState(jwtSecret []byte, state string) (string, error) {
	token, err := jwt.ParseWithClaims(state, &StateClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid state: %w", err)
	}

	claims, ok := token.Claims.(*StateClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid state claims")
	}

	return claims.Redirect, nil
}

func GetDiscordUser(ctx context.Context, cfg *oauth2.Config, code string) (*DiscordUser, error) {
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discord user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode discord user: %w", err)
	}

	if user.ID == "" {
		return nil, fmt.Errorf("discord returned empty user ID")
	}

	return &user, nil
}
