package models

import (
	"testing"
	"time"
)

func TestContestGetStatus_NotActive(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
		IsActive:  false,
	}
	if got := c.GetStatus(); got != ContestStatusNotStarted {
		t.Errorf("GetStatus() = %s, want %s", got, ContestStatusNotStarted)
	}
}

func TestContestGetStatus_NotStarted(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		IsActive:  true,
	}
	if got := c.GetStatus(); got != ContestStatusNotStarted {
		t.Errorf("GetStatus() = %s, want %s", got, ContestStatusNotStarted)
	}
}

func TestContestGetStatus_Running(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
		IsActive:  true,
	}
	if got := c.GetStatus(); got != ContestStatusRunning {
		t.Errorf("GetStatus() = %s, want %s", got, ContestStatusRunning)
	}
}

func TestContestGetStatus_Ended(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour),
		IsActive:  true,
	}
	if got := c.GetStatus(); got != ContestStatusEnded {
		t.Errorf("GetStatus() = %s, want %s", got, ContestStatusEnded)
	}
}

func TestIsValidNotificationType(t *testing.T) {
	tests := []struct {
		ntype string
		want  bool
	}{
		{"info", true},
		{"warning", true},
		{"success", true},
		{"error", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ntype, func(t *testing.T) {
			if got := IsValidNotificationType(tt.ntype); got != tt.want {
				t.Errorf("IsValidNotificationType(%q) = %v, want %v", tt.ntype, got, tt.want)
			}
		})
	}
}
