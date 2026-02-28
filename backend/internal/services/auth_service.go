package services

import (
	"errors"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repositories.UserRepository
	emailService *EmailService
	config       *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, emailService *EmailService, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
		config:       cfg,
	}
}

// Register creates a new user account with email verification
func (s *AuthService) Register(username, email, password string) error {
	// Registration access control
	switch s.config.RegistrationMode {
	case "disabled":
		return errors.New("registration is currently closed")
	case "domain":
		allowed := false
		for _, domain := range strings.Split(s.config.RegistrationAllowedDomains, ",") {
			if strings.HasSuffix(email, "@"+strings.TrimSpace(domain)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New("registration is restricted to approved email domains")
		}
	}

	// Validate email format and domain
	if err := s.emailService.ValidateEmail(email); err != nil {
		return err
	}

	// Check if username already exists
	existingUser, _ := s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// Check if email already exists
	existingEmail, _ := s.userRepo.FindByEmail(email)
	if existingEmail != nil {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate verification token
	token, err := s.emailService.GenerateVerificationToken()
	if err != nil {
		return err
	}

	// Create user
	user := &models.User{
		Username:           username,
		Email:              email,
		PasswordHash:       string(hashedPassword),
		Role:               "user",
		EmailVerified:      false,
		VerificationToken:  token,
		VerificationExpiry: s.emailService.GetVerificationExpiry().Format(time.RFC3339),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(email, username, token); err != nil {
		// Log error but don't fail registration
		// User can request a new verification email later
		return errors.New("account created but failed to send verification email - please request a new one")
	}

	return nil
}

// VerifyEmail verifies a user's email using the verification token
func (s *AuthService) VerifyEmail(token string) error {
	user, err := s.userRepo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	t, _ := time.Parse(time.RFC3339, user.VerificationExpiry)
	if time.Now().After(t) {
		return errors.New("verification token has expired")
	}

	// Mark email as verified
	user.EmailVerified = true
	user.VerificationToken = ""
	user.UpdatedAt = time.Now()

	return s.userRepo.UpdateUser(user)
}

// ResendVerificationEmail sends a new verification email
func (s *AuthService) ResendVerificationEmail(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("email not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	// Generate new token
	token, err := s.emailService.GenerateVerificationToken()
	if err != nil {
		return err
	}

	user.VerificationToken = token
	user.VerificationExpiry = s.emailService.GetVerificationExpiry().Format(time.RFC3339)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return err
	}

	return s.emailService.SendVerificationEmail(user.Email, user.Username, token)
}

// UserInfo contains basic user information returned on login
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login authenticates a user and returns a JWT token and user info
func (s *AuthService) Login(usernameOrEmail, password string) (string, *UserInfo, error) {
	// Try to find by username first
	user, err := s.userRepo.FindByUsername(usernameOrEmail)
	if err != nil {
		// Try to find by email
		user, err = s.userRepo.FindByEmail(usernameOrEmail)
		if err != nil {
			return "", nil, errors.New("invalid credentials")
		}
	}

	// Check if email is verified
	if !user.EmailVerified {
		return "", nil, errors.New("please verify your email before logging in")
	}

	// Check if user is banned
	if user.Status == "banned" {
		return "", nil, errors.New("your account has been banned")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	return s.GenerateToken(user)
}

// GenerateToken generates a JWT token and UserInfo for a user
func (s *AuthService) GenerateToken(user *models.User) (string, *UserInfo, error) {
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", nil, err
	}

	// Create user info response
	userInfo := &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return tokenString, userInfo, nil
}

// GetUserInfo returns user info for a given user ID
func (s *AuthService) GetUserInfo(userID string) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// RecordUserLoginIP records the user's IP address after a successful login
func (s *AuthService) RecordUserLoginIP(userID string, ip string) error {
	return s.userRepo.RecordUserIP(userID, ip, "login")
}

// RequestPasswordReset sends a password reset email
func (s *AuthService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate reset token
	token, err := s.emailService.GenerateVerificationToken()
	if err != nil {
		return err
	}

	user.ResetPasswordToken = token
	user.ResetPasswordExpiry = s.emailService.GetResetPasswordExpiry().Format(time.RFC3339)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return err
	}

	return s.emailService.SendPasswordResetEmail(user.Email, user.Username, token)
}

// ResetPassword resets a user's password using the reset token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	user, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	t, _ := time.Parse(time.RFC3339, user.ResetPasswordExpiry)
	if time.Now().After(t) {
		return errors.New("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.ResetPasswordToken = ""
	user.UpdatedAt = time.Now()

	return s.userRepo.UpdateUser(user)
}

// ChangePassword allows a logged-in user to change their password
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.UpdateUser(user)
}

// UpdateUsername allows a logged-in user to change their username
func (s *AuthService) UpdateUsername(userID, newUsername string) error {
	// Trim whitespace
	newUsername = strings.TrimSpace(newUsername)
	if len(newUsername) < 3 || len(newUsername) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Username == newUsername {
		return nil // No change needed
	}

	// Check if new username is already taken
	existingUser, _ := s.userRepo.FindByUsername(newUsername)
	if existingUser != nil {
		return errors.New("username already exists")
	}

	user.Username = newUsername
	user.UpdatedAt = time.Now()

	return s.userRepo.UpdateUser(user)
}

// GetEnvironment returns the current application environment
func (s *AuthService) GetEnvironment() string {
	return s.config.Environment
}
