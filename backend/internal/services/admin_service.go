package services

import (
	"errors"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	userRepo *repositories.UserRepository
}

func NewAdminService(userRepo *repositories.UserRepository) *AdminService {
	return &AdminService{
		userRepo: userRepo,
	}
}

// CreateAdminUser creates a new admin user with email already verified
func (s *AdminService) CreateAdminUser(username, email, password string) error {
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

	// Create admin user
	user := &models.User{
		ID:            primitive.NewObjectID(),
		Username:      username,
		Email:         email,
		PasswordHash:  string(hashedPassword),
		Role:          "admin",
		EmailVerified: true, // Auto-verify admin users
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.userRepo.CreateUser(user)
}

// PromoteToAdmin promotes an existing user to admin role
func (s *AdminService) PromoteToAdmin(usernameOrEmail string) (*models.User, error) {
	// Try to find by username first
	user, err := s.userRepo.FindByUsername(usernameOrEmail)
	if err != nil {
		// Try to find by email
		user, err = s.userRepo.FindByEmail(usernameOrEmail)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}

	if user.Role == "admin" {
		return nil, errors.New("user is already an admin")
	}

	user.Role = "admin"
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DemoteToUser demotes an admin to regular user role
func (s *AdminService) DemoteToUser(usernameOrEmail string) (*models.User, error) {
	// Try to find by username first
	user, err := s.userRepo.FindByUsername(usernameOrEmail)
	if err != nil {
		// Try to find by email
		user, err = s.userRepo.FindByEmail(usernameOrEmail)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}

	if user.Role == "user" {
		return nil, errors.New("user is already a regular user")
	}

	user.Role = "user"
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers returns all users in the system
func (s *AdminService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAllUsers()
}

// FindUser finds a user by username or email
func (s *AdminService) FindUser(usernameOrEmail string) (*models.User, error) {
	// Try to find by username first
	user, err := s.userRepo.FindByUsername(usernameOrEmail)
	if err != nil {
		// Try to find by email
		user, err = s.userRepo.FindByEmail(usernameOrEmail)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}
	return user, nil
}
