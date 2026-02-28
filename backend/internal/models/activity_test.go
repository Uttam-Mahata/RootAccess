package models

import (
	"testing"
	"github.com/google/uuid"
)

func TestUserActivity(t *testing.T) {
	activity := UserActivity{
		UserID:      uuid.New().String(),
		Username:    "testuser",
		TotalSolves: 5,
		TotalPoints: 500,
		Rank:        3,
		SolveStreak: 2,
	}

	if activity.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", activity.Username)
	}
	if activity.TotalSolves != 5 {
		t.Errorf("Expected 5 solves, got %d", activity.TotalSolves)
	}
	if activity.Rank != 3 {
		t.Errorf("Expected rank 3, got %d", activity.Rank)
	}
}

func TestAdminAnalytics(t *testing.T) {
	analytics := AdminAnalytics{
		TotalUsers:       100,
		TotalTeams:       25,
		TotalChallenges:  50,
		TotalSubmissions: 500,
		TotalCorrect:     200,
		SuccessRate:      0.4,
	}

	if analytics.TotalUsers != 100 {
		t.Errorf("Expected 100 users, got %d", analytics.TotalUsers)
	}
	if analytics.SuccessRate != 0.4 {
		t.Errorf("Expected success rate 0.4, got %f", analytics.SuccessRate)
	}
}

func TestCategoryStat(t *testing.T) {
	stat := CategoryStat{
		Total:  10,
		Solved: 7,
		Points: 350,
	}

	if stat.Total != 10 {
		t.Errorf("Expected total 10, got %d", stat.Total)
	}
	if stat.Solved != 7 {
		t.Errorf("Expected solved 7, got %d", stat.Solved)
	}
}
