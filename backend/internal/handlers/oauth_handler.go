package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

type OAuthHandler struct {
	oauthService *services.OAuthService
	redisClient  *redis.Client
	config       *config.Config
}

func NewOAuthHandler(oauthService *services.OAuthService, redisClient *redis.Client, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		redisClient:  redisClient,
		config:       cfg,
	}
}

// generateStateToken creates a random CSRF state token
func (h *OAuthHandler) generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GoogleLogin initiates the Google OAuth flow
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	// Generate CSRF state token
	state, err := h.generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state token"})
		return
	}

	// Store state in Redis with 10 minute expiry
	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := h.redisClient.Set(ctx, stateKey, "valid", 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store state token"})
		return
	}

	// Get Google OAuth URL
	authURL := h.oauthService.GetGoogleAuthURL(state)

	// Redirect to Google consent page
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles the OAuth callback from Google
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	// Get state and code from query params
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	// Validate state token (CSRF protection)
	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	val, err := h.redisClient.Get(ctx, stateKey).Result()
	if err != nil || val != "valid" {
		h.redirectToFrontendWithError(c, "invalid or expired state token")
		return
	}

	// Delete the state token (one-time use)
	h.redisClient.Del(ctx, stateKey)

	// Handle Google callback
	token, userInfo, err := h.oauthService.HandleGoogleCallback(ctx, code)
	if err != nil {
		h.redirectToFrontendWithError(c, fmt.Sprintf("OAuth authentication failed: %s", err.Error()))
		return
	}

	// Set JWT token in HTTP-only cookie
	c.SetCookie(
		"auth_token",   // name
		token,          // value
		7*24*60*60,     // maxAge (7 days in seconds)
		"/",            // path
		"",             // domain (empty = current domain)
		false,          // secure (set to true in production with HTTPS)
		true,           // httpOnly
	)

	// Redirect to frontend success page
	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// GitHubLogin initiates the GitHub OAuth flow
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	state, err := h.generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state token"})
		return
	}

	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := h.redisClient.Set(ctx, stateKey, "valid", 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store state token"})
		return
	}

	authURL := h.oauthService.GetGitHubAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubCallback handles the OAuth callback from GitHub
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	val, err := h.redisClient.Get(ctx, stateKey).Result()
	if err != nil || val != "valid" {
		h.redirectToFrontendWithError(c, "invalid or expired state token")
		return
	}

	h.redisClient.Del(ctx, stateKey)

	token, userInfo, err := h.oauthService.HandleGitHubCallback(ctx, code)
	if err != nil {
		h.redirectToFrontendWithError(c, fmt.Sprintf("OAuth authentication failed: %s", err.Error()))
		return
	}

	c.SetCookie("auth_token", token, 7*24*60*60, "/", "", false, true)

	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// DiscordLogin initiates the Discord OAuth flow
func (h *OAuthHandler) DiscordLogin(c *gin.Context) {
	state, err := h.generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state token"})
		return
	}

	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := h.redisClient.Set(ctx, stateKey, "valid", 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store state token"})
		return
	}

	authURL := h.oauthService.GetDiscordAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// DiscordCallback handles the OAuth callback from Discord
func (h *OAuthHandler) DiscordCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	val, err := h.redisClient.Get(ctx, stateKey).Result()
	if err != nil || val != "valid" {
		h.redirectToFrontendWithError(c, "invalid or expired state token")
		return
	}

	h.redisClient.Del(ctx, stateKey)

	token, userInfo, err := h.oauthService.HandleDiscordCallback(ctx, code)
	if err != nil {
		h.redirectToFrontendWithError(c, fmt.Sprintf("OAuth authentication failed: %s", err.Error()))
		return
	}

	c.SetCookie("auth_token", token, 7*24*60*60, "/", "", false, true)

	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// redirectToFrontendWithError redirects to frontend with error message
func (h *OAuthHandler) redirectToFrontendWithError(c *gin.Context, errorMsg string) {
	errorURL := fmt.Sprintf("%s/login?error=%s", h.config.FrontendURL, errorMsg)
	c.Redirect(http.StatusTemporaryRedirect, errorURL)
}
