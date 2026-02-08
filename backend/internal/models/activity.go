package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserActivity represents aggregated user activity stats
type UserActivity struct {
	UserID           primitive.ObjectID      `json:"user_id"`
	Username         string                  `json:"username"`
	TotalSolves      int                     `json:"total_solves"`
	TotalPoints      int                     `json:"total_points"`
	CategoryProgress map[string]CategoryStat `json:"category_progress"`
	RecentSolves     []SolveEntry            `json:"recent_solves"`
	Achievements     []Achievement           `json:"achievements"`
	Rank             int                     `json:"rank"`
	SolveStreak      int                     `json:"solve_streak"`
}

// CategoryStat represents solve stats for a category
type CategoryStat struct {
	Total  int `json:"total"`
	Solved int `json:"solved"`
	Points int `json:"points"`
}

// SolveEntry represents a single solve event
type SolveEntry struct {
	ChallengeID    primitive.ObjectID `json:"challenge_id"`
	ChallengeTitle string             `json:"challenge_title"`
	Category       string             `json:"category"`
	Points         int                `json:"points"`
	SolvedAt       time.Time          `json:"solved_at"`
}

// AdminAnalytics represents platform-wide analytics
type AdminAnalytics struct {
	TotalUsers          int                   `json:"total_users"`
	TotalTeams          int                   `json:"total_teams"`
	TotalChallenges     int                   `json:"total_challenges"`
	TotalSubmissions    int                   `json:"total_submissions"`
	TotalCorrect        int                   `json:"total_correct"`
	SuccessRate         float64               `json:"success_rate"`
	ChallengePopularity []ChallengePopularity `json:"challenge_popularity"`
	CategoryBreakdown   map[string]int        `json:"category_breakdown"`
	DifficultyBreakdown map[string]int        `json:"difficulty_breakdown"`
	RecentActivity      []RecentActivityEntry `json:"recent_activity"`
	SolvesOverTime      []TimeSeriesEntry     `json:"solves_over_time"`
}

// ChallengePopularity represents a challenge with its solve metrics
type ChallengePopularity struct {
	ChallengeID  primitive.ObjectID `json:"challenge_id"`
	Title        string             `json:"title"`
	Category     string             `json:"category"`
	SolveCount   int                `json:"solve_count"`
	AttemptCount int                `json:"attempt_count"`
	SuccessRate  float64            `json:"success_rate"`
}

// RecentActivityEntry represents a recent platform activity event
type RecentActivityEntry struct {
	UserID        primitive.ObjectID `json:"user_id"`
	Username      string             `json:"username"`
	Action        string             `json:"action"`
	ChallengeID   primitive.ObjectID `json:"challenge_id,omitempty"`
	ChallengeName string             `json:"challenge_name,omitempty"`
	Timestamp     time.Time          `json:"timestamp"`
}

// TimeSeriesEntry represents a data point in a time series
type TimeSeriesEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
