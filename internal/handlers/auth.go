package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"eva_trainer/internal/middleware"
	"eva_trainer/internal/models"
	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	userRepo        models.UserRepository
	refreshRepo     models.RefreshTokenRepository
	oauthCfg        *oauth2.Config
	jwtService      *services.JWTService
	frontendBaseURL string
	logger          services.Logger
}

func NewAuthHandler(
	userRepo models.UserRepository,
	refreshRepo models.RefreshTokenRepository,
	oauthCfg *oauth2.Config,
	jwtService *services.JWTService,
	frontendBaseURL string,
	logger services.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		oauthCfg:        oauthCfg,
		jwtService:      jwtService,
		frontendBaseURL: frontendBaseURL,
		logger:          logger,
	}
}

func (h *AuthHandler) LoginRedirect(c *gin.Context) {
	redirect := c.DefaultQuery("redirect", "/")

	state, err := services.GenerateOAuthState(h.jwtService.SecretBytes(), redirect)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_code", "authorization code is required", http.StatusBadRequest))
		return
	}

	if state == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("missing_state", "state parameter is required", http.StatusBadRequest))
		return
	}

	redirectPath, err := services.ValidateOAuthState(h.jwtService.SecretBytes(), state)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_state", "invalid state parameter", http.StatusBadRequest))
		return
	}

	discordUser, err := services.GetDiscordUser(c.Request.Context(), h.oauthCfg, code)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			apierror.RespondWithError(c, apierror.NewAPIError(
				apierror.ErrCodeDiscordAPI,
				"Discord est actuellement indisponible, veuillez réessayer",
				http.StatusBadGateway,
			))
		} else {
			apierror.RespondWithError(c, apierror.NewAPIError(
				apierror.ErrCodeDiscordAPI,
				"connexion impossible",
				http.StatusBadGateway,
			))
		}
		return
	}

	user, isNew, err := h.userRepo.FindOrCreateByDiscordID(discordUser.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	adminCount, _ := h.userRepo.AdminCount()
	if adminCount == 0 {
		h.userRepo.SetAdmin(user.ID)
	}

	if isNew {
		displayName := discordUser.GlobalName
		if displayName == "" {
			displayName = discordUser.Username
		}
		if displayName != "" {
			h.userRepo.UpdateDisplayName(user.ID, displayName)
		}
	}

	if discordUser.Username != "" {
		h.userRepo.SetDiscordUsername(user.ID, discordUser.Username)
	}

	accessToken, expiresAt, err := h.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	refreshPlain, err := h.refreshRepo.CreateRefreshToken(user.ID, time.Now().Add(h.jwtService.RefreshExpiry()))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
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
		apierror.RespondWithError(c, apierror.NewAPIError("missing_refresh_token", "refresh_token is required", http.StatusBadRequest))
		return
	}

	tokenHash := hashToken(req.RefreshToken)

	existing, err := h.refreshRepo.FindRefreshTokenByHash(tokenHash)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("refresh_token_not_found", "invalid refresh token", http.StatusUnauthorized))
		return
	}

	if existing.RevokedAt != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("refresh_token_revoked", "refresh token has been revoked", http.StatusUnauthorized))
		return
	}

	if time.Now().After(existing.ExpiresAt) {
		apierror.RespondWithError(c, apierror.NewAPIError("refresh_token_expired", "refresh token has expired", http.StatusUnauthorized))
		return
	}

	newPlain, err := h.refreshRepo.RotateRefreshToken(tokenHash, time.Now().Add(h.jwtService.RefreshExpiry()))
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	accessToken, expiresAt, err := h.jwtService.GenerateAccessToken(existing.UserID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newPlain,
		"expires_in":    int(time.Until(expiresAt).Seconds()),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"id":               user.ID,
		"discord_id":       user.DiscordID,
		"discord_username": user.DiscordUsername,
		"display_name":     user.DisplayName,
		"is_admin":         user.IsAdmin,
		"created_at":       user.CreatedAt,
	})
}

func (h *AuthHandler) UpdateDisplayName(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("unauthorized", "unauthorized", http.StatusUnauthorized))
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.DisplayName == "" {
		apierror.RespondWithError(c, apierror.NewAPIError("invalid_display_name", "display name is required", http.StatusBadRequest))
		return
	}

	if err := h.userRepo.UpdateDisplayName(user.ID, req.DisplayName); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError("internal_error", "internal server error", http.StatusInternalServerError))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"display_name": req.DisplayName,
	})
}

func (h *AuthHandler) GetSettings(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	settings, err := h.userRepo.GetUserSettings(user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError(
			"internal_error", "failed to load settings", http.StatusInternalServerError,
		))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"revision_correct_timer":   settings.RevisionCorrectTimer,
		"revision_incorrect_timer": settings.RevisionIncorrectTimer,
		"revision_question_count":  settings.RevisionQuestionCount,
		"revision_review_zones":    settings.RevisionReviewZones,
		"revision_review_lines":    settings.RevisionReviewLines,
	})
}

func (h *AuthHandler) UpdateSettings(c *gin.Context) {
	user, err := middleware.MustGetUser(c)
	if err != nil {
		return
	}

	var req struct {
		RevisionCorrectTimer   *int  `json:"revision_correct_timer"`
		RevisionIncorrectTimer *int  `json:"revision_incorrect_timer"`
		RevisionQuestionCount  *int  `json:"revision_question_count"`
		RevisionReviewZones    *bool `json:"revision_review_zones"`
		RevisionReviewLines    *bool `json:"revision_review_lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError(
			"invalid_settings", "invalid settings format", http.StatusBadRequest,
		))
		return
	}

	if (req.RevisionCorrectTimer != nil && *req.RevisionCorrectTimer < 1) ||
		(req.RevisionIncorrectTimer != nil && *req.RevisionIncorrectTimer < 1) ||
		(req.RevisionQuestionCount != nil && *req.RevisionQuestionCount < 1) {
		apierror.RespondWithError(c, apierror.NewAPIError(
			"invalid_settings", "values must be >= 1", http.StatusBadRequest,
		))
		return
	}

	current, err := h.userRepo.GetUserSettings(user.ID)
	if err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError(
			"internal_error", "failed to load settings", http.StatusInternalServerError,
		))
		return
	}

	merged := *current
	if req.RevisionCorrectTimer != nil {
		merged.RevisionCorrectTimer = *req.RevisionCorrectTimer
	}
	if req.RevisionIncorrectTimer != nil {
		merged.RevisionIncorrectTimer = *req.RevisionIncorrectTimer
	}
	if req.RevisionQuestionCount != nil {
		merged.RevisionQuestionCount = *req.RevisionQuestionCount
	}
	if req.RevisionReviewZones != nil {
		merged.RevisionReviewZones = *req.RevisionReviewZones
	}
	if req.RevisionReviewLines != nil {
		merged.RevisionReviewLines = *req.RevisionReviewLines
	}

	if err := h.userRepo.UpdateUserSettings(user.ID, &merged); err != nil {
		apierror.RespondWithError(c, apierror.NewAPIError(
			"internal_error", "failed to save settings", http.StatusInternalServerError,
		))
		return
	}

	apierror.RespondWithSuccess(c, http.StatusOK, gin.H{
		"revision_correct_timer":   merged.RevisionCorrectTimer,
		"revision_incorrect_timer": merged.RevisionIncorrectTimer,
		"revision_question_count":  merged.RevisionQuestionCount,
		"revision_review_zones":    merged.RevisionReviewZones,
		"revision_review_lines":    merged.RevisionReviewLines,
	})
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
