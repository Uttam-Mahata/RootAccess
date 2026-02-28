package handlers

import (
	"net/http"
	"strings"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// Register creates a new user account and sends verification email
// @Summary Register a new user
// @Description Create a new user account with username, email, and password. Role is hardcoded to "user".
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := h.authService.Register(req.Username, req.Email, req.Password); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful! Please check your email to verify your account.",
	})
}

// VerifyEmail verifies user's email with token
// @Summary Verify email address
// @Description Verify a user's email address using the token sent during registration.
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string false "Verification token"
// @Param request body VerifyEmailRequest false "Verification token in body"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/verify-email [get]
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Verification token is required", err)
			return
		}
		token = req.Token
	}

	if err := h.authService.VerifyEmail(token); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully! You can now log in.",
	})
}

// ResendVerification sends a new verification email
// @Summary Resend verification email
// @Description Send a new verification email to the user.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "Email details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := h.authService.ResendVerificationEmail(req.Email); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully!",
	})
}

// Login authenticates a user and sets JWT token in HTTP-only cookie
// @Summary Login user
// @Description Authenticate a user and set a JWT token in an HTTP-only cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	token, userInfo, err := h.authService.Login(req.UsernameOrEmail, req.Password)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err.Error(), err)
		return
	}

	// Record user IP address after successful login
	clientIP := c.ClientIP()
	go h.authService.RecordUserLoginIP(userInfo.ID, clientIP)

	// Set HTTP-only cookie with JWT token
	isProd := h.authService.GetEnvironment() == "production"

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

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful!",
		"user":    userInfo,
	})
}

// Logout clears the authentication cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	isProd := h.authService.GetEnvironment() == "production"

	// Clear the cookie by setting maxAge to -1
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		isProd,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully!",
	})
}

// ForgotPassword initiates password reset process
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		// Don't reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{
			"message": "If your email is registered, you will receive a password reset link.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If your email is registered, you will receive a password reset link.",
	})
}

// ResetPassword resets user password using reset token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully! You can now log in with your new password.",
	})
}

// ChangePassword allows logged-in user to change password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	if err := h.authService.ChangePassword(userID.(string), req.OldPassword, req.NewPassword); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully!",
	})
}

// UpdateUsername allows logged-in user to change their username
func (h *AuthHandler) UpdateUsername(c *gin.Context) {
	var req UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	if err := h.authService.UpdateUsername(userID.(string), req.Username); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Update the JWT cookie with the new username
	user, err := h.authService.GetUserInfo(userID.(string))
	if err == nil {
		// Generate new token
		token, _, err := h.authService.GenerateToken(user)
		if err == nil {
			isProd := h.authService.GetEnvironment() == "production"
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(
				"auth_token",
				token,
				7*24*60*60,
				"/",
				"",
				isProd,
				true,
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Username updated successfully!",
	})
}

// GetMe returns current user info when authenticated via cookie or Bearer token.
// Records last IP and last login on each successful check so admin sees up-to-date activity.
func (h *AuthHandler) GetMe(c *gin.Context) {
	tokenString, err := c.Cookie("auth_token")
	if err != nil || tokenString == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.authService.GetJWTSecret()), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	userID, _ := claims["user_id"].(string)
	if userID != "" {
		go h.authService.RecordUserLoginIP(userID, utils.GetClientIP(c))
	}
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       claims["user_id"],
			"username": claims["username"],
			"email":    claims["email"],
			"role":     claims["role"],
		},
	})
}

// GetToken returns the current JWT token from the cookie
// @Summary Get current JWT token
// @Description Retrieve the current authenticated JWT token. Used by the CLI for seamless login.
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/token [get]
func (h *AuthHandler) GetToken(c *gin.Context) {
	token, err := c.Cookie("auth_token")
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
