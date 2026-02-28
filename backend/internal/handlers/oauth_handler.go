package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/gin-gonic/gin"
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
// @Summary Google OAuth login
// @Description Redirect to Google's OAuth consent page. Supports 'cli=true' for CLI-based login.
// @Tags Auth
// @Param cli query bool false "Set to true for CLI-based login"
// @Success 307 {string} string "Redirect to Google"
// @Router /auth/google [get]
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	if h.redisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth session storage is unavailable. Please check backend logs."})
		return
	}

	// Generate CSRF state token
	state, err := h.generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state token"})
		return
	}

	// Check if this is a CLI login
	isCLI := c.Query("cli") == "true"
	stateValue := "valid"
	if isCLI {
		stateValue = "cli"
	}

	// Store state in Redis with 10 minute expiry
	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	if err := h.redisClient.Set(ctx, stateKey, stateValue, 10*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store state token"})
		return
	}

	// Get Google OAuth URL
	authURL := h.oauthService.GetGoogleAuthURL(state)

	// Redirect to Google consent page
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles the OAuth callback from Google
// @Summary Google OAuth callback
// @Description Callback endpoint for Google OAuth. Validates state and code, then sets JWT cookie.
// @Tags Auth
// @Param state query string true "CSRF state token"
// @Param code query string true "Authorization code"
// @Success 307 {string} string "Redirect to Frontend"
// @Router /auth/google/callback [get]
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	// Get state and code from query params
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	if h.redisClient == nil {
		h.redirectToFrontendWithError(c, "OAuth session storage is unavailable")
		return
	}

	// Validate state token (CSRF protection)
	ctx := context.Background()
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	val, err := h.redisClient.Get(ctx, stateKey).Result()
	if err != nil || (val != "valid" && val != "cli") {
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

	// Record last IP and last login (non-blocking)
	go h.oauthService.RecordUserLoginIP(userInfo.ID, c.ClientIP())

	// If it's a CLI login, return the token in a simple HTML page
	if val == "cli" {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>RootAccess CLI Login Success</title>
				<style>
					body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background-color: #1a202c; color: #e2e8f0; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
					.card { background-color: #2d3748; padding: 2.5rem; border-radius: 0.75rem; box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3); max-width: 28rem; width: 100%%; text-align: center; }
					h1 { color: #63b3ed; margin-bottom: 1.5rem; }
					p { margin-bottom: 2rem; color: #a0aec0; line-height: 1.6; }
					.token-container { background-color: #1a202c; padding: 1rem; border-radius: 0.5rem; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; font-size: 0.875rem; word-break: break-all; margin-bottom: 2rem; border: 1px solid #4a5568; }
					button { background-color: #4299e1; color: white; border: none; padding: 0.75rem 1.5rem; border-radius: 0.375rem; font-weight: 600; cursor: pointer; transition: background-color 0.2s; }
					button:hover { background-color: #3182ce; }
				</style>
			</head>
			<body>
				<div class="card">
					<h1>RootAccess CLI</h1>
					<p>Login successful! Copy the token below and paste it into your CLI terminal.</p>
					<div class="token-container" id="token">%s</div>
					<button onclick="copyToken()">Copy to Clipboard</button>
				</div>
				<script>
					function copyToken() {
						const token = document.getElementById('token').innerText;
						navigator.clipboard.writeText(token).then(() => {
							const btn = document.querySelector('button');
							btn.innerText = 'Copied!';
							btn.style.backgroundColor = '#48bb78';
							setTimeout(() => {
								btn.innerText = 'Copy to Clipboard';
								btn.style.backgroundColor = '#4299e1';
							}, 2000);
						});
					}
				</script>
			</body>
			</html>
		`, token))
		return
	}

	// Set JWT token in HTTP-only cookie
	isProd := h.config.Environment == "production"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token", // name
		token,        // value
		7*24*60*60,   // maxAge (7 days in seconds)
		"/",          // path
		"",           // domain (empty = current domain)
		isProd,       // secure (set to true in production with HTTPS)
		true,         // httpOnly
	)

	// Redirect to frontend success page
	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// GitHubLogin initiates the GitHub OAuth flow
// @Summary GitHub OAuth login
// @Description Redirect to GitHub's OAuth consent page.
// @Tags Auth
// @Success 307 {string} string "Redirect to GitHub"
// @Router /auth/github [get]
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	if h.redisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth session storage is unavailable"})
		return
	}

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
// @Summary GitHub OAuth callback
// @Description Callback endpoint for GitHub OAuth.
// @Tags Auth
// @Param state query string true "CSRF state token"
// @Param code query string true "Authorization code"
// @Success 307 {string} string "Redirect to Frontend"
// @Router /auth/github/callback [get]
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	if h.redisClient == nil {
		h.redirectToFrontendWithError(c, "OAuth session storage is unavailable")
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

	go h.oauthService.RecordUserLoginIP(userInfo.ID, c.ClientIP())

	isProd := h.config.Environment == "production"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 7*24*60*60, "/", "", isProd, true)

	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)

	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// DiscordLogin initiates the Discord OAuth flow
// @Summary Discord OAuth login
// @Description Redirect to Discord's OAuth consent page.
// @Tags Auth
// @Success 307 {string} string "Redirect to Discord"
// @Router /auth/discord [get]
func (h *OAuthHandler) DiscordLogin(c *gin.Context) {
	if h.redisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth session storage is unavailable"})
		return
	}

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
// @Summary Discord OAuth callback
// @Description Callback endpoint for Discord OAuth.
// @Tags Auth
// @Param state query string true "CSRF state token"
// @Param code query string true "Authorization code"
// @Success 307 {string} string "Redirect to Frontend"
// @Router /auth/discord/callback [get]
func (h *OAuthHandler) DiscordCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		h.redirectToFrontendWithError(c, "missing state or code parameter")
		return
	}

	if h.redisClient == nil {
		h.redirectToFrontendWithError(c, "OAuth session storage is unavailable")
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

	go h.oauthService.RecordUserLoginIP(userInfo.ID, c.ClientIP())

	isProd := h.config.Environment == "production"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 7*24*60*60, "/", "", isProd, true)

	successURL := fmt.Sprintf("%s/auth/callback?success=true&username=%s", h.config.FrontendURL, userInfo.Username)

	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// redirectToFrontendWithError redirects to frontend with error message
func (h *OAuthHandler) redirectToFrontendWithError(c *gin.Context, errorMsg string) {
	errorURL := fmt.Sprintf("%s/login?error=%s", h.config.FrontendURL, errorMsg)
	c.Redirect(http.StatusTemporaryRedirect, errorURL)
}
