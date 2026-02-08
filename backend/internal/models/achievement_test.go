package models

import (
	"testing"
)

func TestAchievementTypes(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FirstBlood", AchievementFirstBlood, "first_blood"},
		{"CategoryMaster", AchievementCategoryMaster, "category_master"},
		{"Streak", AchievementStreak, "streak"},
		{"SpeedDemon", AchievementSpeedDemon, "speed_demon"},
		{"NightOwl", AchievementNightOwl, "night_owl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}
