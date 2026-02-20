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
	RecentActivity       []RecentActivityEntry `json:"recent_activity"`
	SolvesOverTime       []TimeSeriesEntry     `json:"solves_over_time"`
	SubmissionsOverTime  []TimeSeriesEntry     `json:"submissions_over_time"`
	// Enhanced statistics
	ActiveUsers         int               `json:"active_users"`
	BannedUsers         int               `json:"banned_users"`
	VerifiedUsers       int               `json:"verified_users"`
	AdminCount          int               `json:"admin_count"`
	NewUsersToday       int               `json:"new_users_today"`
	NewUsersThisWeek    int               `json:"new_users_this_week"`
	NewTeamsToday       int               `json:"new_teams_today"`
	NewTeamsThisWeek    int               `json:"new_teams_this_week"`
	SubmissionsToday    int               `json:"submissions_today"`
	SolvesToday         int               `json:"solves_today"`
	AverageTeamSize     float64           `json:"average_team_size"`
	UserGrowth          []TimeSeriesEntry `json:"user_growth"`
	TeamGrowth          []TimeSeriesEntry `json:"team_growth"`
	TopTeams            []TeamStats       `json:"top_teams"`
	TopUsers            []UserStats       `json:"top_users"`
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

// TeamStats represents team statistics for leaderboard
type TeamStats struct {
	TeamID      primitive.ObjectID `json:"team_id"`
	Name        string             `json:"name"`
	Score       int                `json:"score"`
	MemberCount int                `json:"member_count"`
	SolveCount  int                `json:"solve_count"`
}

// UserStats represents user statistics for leaderboard
type UserStats struct {
	UserID     primitive.ObjectID `json:"user_id"`
	Username   string             `json:"username"`
	Score      int                `json:"score"`
	SolveCount int                `json:"solve_count"`
}
