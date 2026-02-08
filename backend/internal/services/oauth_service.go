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

var githubEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

type OAuthService struct {
	userRepo      *repositories.UserRepository
	config        *config.Config
	googleConfig  *oauth2.Config
	githubConfig  *oauth2.Config
	discordConfig *oauth2.Config
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

type GitHubUserInfo struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type DiscordUserInfo struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Discriminator string `json:"discriminator"`
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

	githubConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     githubEndpoint,
	}

	discordConfig := &oauth2.Config{
		ClientID:     cfg.DiscordClientID,
		ClientSecret: cfg.DiscordClientSecret,
		RedirectURL:  cfg.DiscordRedirectURL,
		Scopes:       []string{"identify", "email"},
		Endpoint:     discordEndpoint,
	}

	return &OAuthService{
		userRepo:      userRepo,
		config:        cfg,
		googleConfig:  googleConfig,
		githubConfig:  githubConfig,
		discordConfig: discordConfig,
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

	username := googleUser.Email
	if googleUser.GivenName != "" {
		username = googleUser.GivenName
	}

	user, err := s.findOrCreateOAuthUser("google", googleUser.ID, googleUser.Email, username, token)
	if err != nil {
		return "", nil, err
	}

	return s.generateJWTAndUserInfo(user)
}

// GetGitHubAuthURL generates the GitHub OAuth consent URL
func (s *OAuthService) GetGitHubAuthURL(state string) string {
	return s.githubConfig.AuthCodeURL(state)
}

// HandleGitHubCallback orchestrates the full GitHub OAuth callback flow
func (s *OAuthService) HandleGitHubCallback(ctx context.Context, code string) (string, *UserInfo, error) {
	token, err := s.githubConfig.Exchange(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	client := s.githubConfig.Client(ctx, token)

	// Fetch user info
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var ghUser GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return "", nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If email is not public, fetch from emails API
	email := ghUser.Email
	if email == "" {
		email, err = s.fetchGitHubPrimaryEmail(ctx, client)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get GitHub email: %w", err)
		}
	}

	username := ghUser.Login
	if username == "" && ghUser.Name != "" {
		username = ghUser.Name
	}

	providerID := fmt.Sprintf("%d", ghUser.ID)
	user, err := s.findOrCreateOAuthUser("github", providerID, email, username, token)
	if err != nil {
		return "", nil, err
	}

	return s.generateJWTAndUserInfo(user)
}

// fetchGitHubPrimaryEmail fetches the primary verified email from GitHub
func (s *OAuthService) fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails API returned status %d", resp.StatusCode)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no verified primary email found on GitHub account")
}

// GetDiscordAuthURL generates the Discord OAuth consent URL
func (s *OAuthService) GetDiscordAuthURL(state string) string {
	return s.discordConfig.AuthCodeURL(state)
}

// HandleDiscordCallback orchestrates the full Discord OAuth callback flow
func (s *OAuthService) HandleDiscordCallback(ctx context.Context, code string) (string, *UserInfo, error) {
	token, err := s.discordConfig.Exchange(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	client := s.discordConfig.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	var discordUser DiscordUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return "", nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if !discordUser.Verified {
		return "", nil, errors.New("discord email not verified")
	}

	if discordUser.Email == "" {
		return "", nil, errors.New("discord account has no email")
	}

	user, err := s.findOrCreateOAuthUser("discord", discordUser.ID, discordUser.Email, discordUser.Username, token)
	if err != nil {
		return "", nil, err
	}

	return s.generateJWTAndUserInfo(user)
}

// findOrCreateOAuthUser finds an existing user by email or creates a new one for OAuth
func (s *OAuthService) findOrCreateOAuthUser(provider, providerID, email, username string, token *oauth2.Token) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// User doesn't exist, create new one
		existingUser, _ := s.userRepo.FindByUsername(username)
		if existingUser != nil {
			username = fmt.Sprintf("%s_%d", username, time.Now().Unix()%10000)
		}

		user = &models.User{
			Username:      username,
			Email:         email,
			PasswordHash:  "",
			Role:          "user",
			EmailVerified: true,
			OAuth: &models.OAuth{
				Provider:     provider,
				ProviderID:   providerID,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				ExpiresAt:    token.Expiry,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.userRepo.CreateUser(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		return user, nil
	}

	// User exists, link or update OAuth
	if user.OAuth == nil || user.OAuth.Provider != provider {
		user.OAuth = &models.OAuth{
			Provider:     provider,
			ProviderID:   providerID,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		}
		user.EmailVerified = true
		user.UpdatedAt = time.Now()

		if err := s.userRepo.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("failed to link OAuth account: %w", err)
		}
	} else {
		user.OAuth.AccessToken = token.AccessToken
		user.OAuth.RefreshToken = token.RefreshToken
		user.OAuth.ExpiresAt = token.Expiry
		user.UpdatedAt = time.Now()

		if err := s.userRepo.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("failed to update OAuth tokens: %w", err)
		}
	}

	return user, nil
}

// generateJWTAndUserInfo creates a JWT token and UserInfo response for a user
func (s *OAuthService) generateJWTAndUserInfo(user *models.User) (string, *UserInfo, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	userInfo := &UserInfo{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return tokenString, userInfo, nil
}

// createOAuthUser creates a new user from Google OAuth data (kept for backward compatibility)
func (s *OAuthService) createOAuthUser(googleUser *GoogleUserInfo, token *oauth2.Token) (*models.User, error) {
	username := googleUser.Email
	if googleUser.GivenName != "" {
		username = googleUser.GivenName
	}
	return s.findOrCreateOAuthUser("google", googleUser.ID, googleUser.Email, username, token)
}
