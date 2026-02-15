package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserSocialLinks(t *testing.T) {
	socialLinks := &SocialLinks{
		GitHub:   "testuser",
		Twitter:  "testhandle",
		Discord:  "testuser#1234",
		LinkedIn: "test-user",
	}

	user := User{
		ID:          primitive.NewObjectID(),
		Username:    "testuser",
		Email:       "test@example.com",
		Role:        "user",
		Bio:         "Security enthusiast",
		Website:     "https://example.com",
		SocialLinks: socialLinks,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if user.Bio != "Security enthusiast" {
		t.Errorf("Expected bio 'Security enthusiast', got '%s'", user.Bio)
	}
	if user.Website != "https://example.com" {
		t.Errorf("Expected website 'https://example.com', got '%s'", user.Website)
	}
	if user.SocialLinks == nil {
		t.Error("Expected social links to be set")
	}
	if user.SocialLinks.GitHub != "testuser" {
		t.Errorf("Expected GitHub 'testuser', got '%s'", user.SocialLinks.GitHub)
	}
	if user.SocialLinks.Twitter != "testhandle" {
		t.Errorf("Expected Twitter 'testhandle', got '%s'", user.SocialLinks.Twitter)
	}
	if user.SocialLinks.Discord != "testuser#1234" {
		t.Errorf("Expected Discord 'testuser#1234', got '%s'", user.SocialLinks.Discord)
	}
	if user.SocialLinks.LinkedIn != "test-user" {
		t.Errorf("Expected LinkedIn 'test-user', got '%s'", user.SocialLinks.LinkedIn)
	}
}

func TestUserWithoutSocialLinks(t *testing.T) {
	user := User{
		ID:       primitive.NewObjectID(),
		Username: "basicuser",
		Email:    "basic@example.com",
		Role:     "user",
		Status:   "active",
	}

	if user.SocialLinks != nil {
		t.Error("Expected social links to be nil for basic user")
	}
	if user.Bio != "" {
		t.Errorf("Expected empty bio, got '%s'", user.Bio)
	}
	if user.Website != "" {
		t.Errorf("Expected empty website, got '%s'", user.Website)
	}
}

func TestIPRecord(t *testing.T) {
	record := IPRecord{
		IP:        "192.168.1.1",
		Timestamp: time.Now(),
		Action:    "login",
	}

	if record.IP != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", record.IP)
	}
	if record.Action != "login" {
		t.Errorf("Expected action 'login', got '%s'", record.Action)
	}
}
