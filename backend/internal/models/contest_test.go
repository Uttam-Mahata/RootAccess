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

func TestContestGetStatus_Paused(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
		IsActive:  true,
		IsPaused:  true,
	}
	if got := c.GetStatus(); got != ContestStatusPaused {
		t.Errorf("GetStatus() = %s, want %s", got, ContestStatusPaused)
	}
}

func TestContestIsScoreboardFrozen_NoFreezeTime(t *testing.T) {
	c := &ContestConfig{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
		IsActive:  true,
	}
	if c.IsScoreboardFrozen() {
		t.Errorf("IsScoreboardFrozen() = true, want false (no freeze time set)")
	}
}

func TestContestIsScoreboardFrozen_BeforeFreezeTime(t *testing.T) {
	ft := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	c := &ContestConfig{
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
		FreezeTime: ft,
		IsActive:   true,
	}
	if c.IsScoreboardFrozen() {
		t.Errorf("IsScoreboardFrozen() = true, want false (before freeze time)")
	}
}

func TestContestIsScoreboardFrozen_AfterFreezeTime(t *testing.T) {
	ft := time.Now().Add(-30 * time.Minute).Format(time.RFC3339)
	c := &ContestConfig{
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
		FreezeTime: ft,
		IsActive:   true,
	}
	if !c.IsScoreboardFrozen() {
		t.Errorf("IsScoreboardFrozen() = false, want true (after freeze time)")
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
