package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	userRepo        models.UserRepository
	refreshRepo     models.RefreshTokenRepository
	oauthCfg        *oauth2.Config
	jwtService      *services.JWTService
	frontendBaseURL string
}

func NewAuthHandler(
	userRepo models.UserRepository,
	refreshRepo models.RefreshTokenRepository,
	oauthCfg *oauth2.Config,
	jwtService *services.JWTService,
	frontendBaseURL string,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		oauthCfg:        oauthCfg,
		jwtService:      jwtService,
		frontendBaseURL: frontendBaseURL,
	}
}

func (h *AuthHandler) LoginRedirect(c *gin.Context) {
	redirect := c.DefaultQuery("redirect", "/")

	state, err := services.GenerateOAuthState(h.jwtService.SecretBytes(), redirect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_code",
			"message": "authorization code is required",
		})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_state",
			"message": "state parameter is required",
		})
		return
	}

	redirectPath, err := services.ValidateOAuthState(h.jwtService.SecretBytes(), state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "invalid state parameter",
		})
		return
	}

	discordID, err := services.GetDiscordUserID(c.Request.Context(), h.oauthCfg, code)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "discord_api_error",
			"message": "connexion impossible",
		})
		return
	}

	user, isNew, err := h.userRepo.FindOrCreateByDiscordID(discordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	if isNew {
		first, err := h.userRepo.IsFirstUser()
		if err == nil && first {
			h.userRepo.SetAdmin(user.ID)
		}
	}

	accessToken, expiresAt, err := h.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	refreshPlain, err := h.refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(h.jwtService.RefreshExpiry()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	expiresIn := int(time.Until(expiresAt).Seconds())
	location := h.frontendBaseURL + "/callback#access_token=" + accessToken +
		"&refresh_token=" + refreshPlain +
		"&expires_in=" + fmt.Sprintf("%d", expiresIn)

	if redirectPath != "" && redirectPath != "/" {
		location += "&redirect=" + redirectPath
	}

	c.Redirect(http.StatusFound, location)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_refresh_token",
			"message": "refresh_token is required",
		})
		return
	}

	tokenHash := hashToken(req.RefreshToken)

	existing, err := h.refreshRepo.FindRefreshTokenByHash(tokenHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "refresh_token_not_found",
			"message": "invalid refresh token",
		})
		return
	}

	if existing.RevokedAt != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "refresh_token_revoked",
			"message": "refresh token has been revoked",
		})
		return
	}

	if time.Now().After(existing.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "refresh_token_expired",
			"message": "refresh token has expired",
		})
		return
	}

	newPlain, err := h.refreshRepo.RotateRefreshToken(tokenHash, time.Now().Add(h.jwtService.RefreshExpiry()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	accessToken, expiresAt, err := h.jwtService.GenerateAccessToken(existing.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newPlain,
		"expires_in":    int(time.Until(expiresAt).Seconds()),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"discord_id": user.DiscordID,
		"is_admin":   user.IsAdmin,
		"created_at": user.CreatedAt,
	})
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
