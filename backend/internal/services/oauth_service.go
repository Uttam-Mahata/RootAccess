package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	userRepo     *repositories.UserRepository
	config       *config.Config
	googleConfig *oauth2.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewOAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *OAuthService {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthService{
		userRepo:     userRepo,
		config:       cfg,
		googleConfig: googleConfig,
	}
}

// GetGoogleAuthURL generates the Google OAuth consent URL with CSRF state token
func (s *OAuthService) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeGoogleCode exchanges authorization code for access token
func (s *OAuthService) ExchangeGoogleCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	return token, nil
}

// GetGoogleUserInfo fetches user profile from Google API
func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// HandleGoogleCallback orchestrates the full OAuth callback flow
func (s *OAuthService) HandleGoogleCallback(ctx context.Context, code string) (string, *UserInfo, error) {
	// Exchange code for token
	token, err := s.ExchangeGoogleCode(ctx, code)
	if err != nil {
		return "", nil, err
	}

	// Get user info from Google
	googleUser, err := s.GetGoogleUserInfo(ctx, token)
	if err != nil {
		return "", nil, err
	}

	if !googleUser.VerifiedEmail {
		return "", nil, errors.New("google email not verified")
	}

	// Find or create user
	user, err := s.userRepo.FindByEmail(googleUser.Email)
	if err != nil {
		// User doesn't exist, create new one
		user, err = s.createOAuthUser(googleUser, token)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, link OAuth account if not already linked
		if user.OAuth == nil || user.OAuth.Provider != "google" {
			user.OAuth = &models.OAuth{
				Provider:     "google",
				ProviderID:   googleUser.ID,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				ExpiresAt:    token.Expiry,
			}
			user.EmailVerified = true
			user.UpdatedAt = time.Now()

			if err := s.userRepo.UpdateUser(user); err != nil {
				return "", nil, fmt.Errorf("failed to link OAuth account: %w", err)
			}
		} else {
			// Update OAuth tokens
			user.OAuth.AccessToken = token.AccessToken
			user.OAuth.RefreshToken = token.RefreshToken
			user.OAuth.ExpiresAt = token.Expiry
			user.UpdatedAt = time.Now()

			if err := s.userRepo.UpdateUser(user); err != nil {
				return "", nil, fmt.Errorf("failed to update OAuth tokens: %w", err)
			}
		}
	}

	// Generate JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := jwtToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Create user info response
	userInfo := &UserInfo{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return tokenString, userInfo, nil
}

// createOAuthUser creates a new user from Google OAuth data
func (s *OAuthService) createOAuthUser(googleUser *GoogleUserInfo, token *oauth2.Token) (*models.User, error) {
	// Generate username from email or name
	username := googleUser.Email
	if googleUser.GivenName != "" {
		username = googleUser.GivenName
	}

	// Ensure username is unique
	existingUser, _ := s.userRepo.FindByUsername(username)
	if existingUser != nil {
		// Append random suffix if username exists
		username = fmt.Sprintf("%s_%d", username, time.Now().Unix()%10000)
	}

	user := &models.User{
		Username:      username,
		Email:         googleUser.Email,
		PasswordHash:  "", // OAuth users don't have password
		Role:          "user",
		EmailVerified: true, // Google already verified the email
		OAuth: &models.OAuth{
			Provider:     "google",
			ProviderID:   googleUser.ID,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}
