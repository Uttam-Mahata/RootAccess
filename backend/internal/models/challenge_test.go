package models

import (
	"testing"
)

func TestCurrentPoints_DynamicScoring(t *testing.T) {
	tests := []struct {
		name       string
		challenge  Challenge
		wantPoints int
	}{
		{
			name: "no solves returns max points",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  0,
				ScoringType: ScoringDynamic,
			},
			wantPoints: 500,
		},
		{
			name: "1 solve returns near max",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  1,
				ScoringType: ScoringDynamic,
			},
			wantPoints: 496,
		},
		{
			name: "decay solves reaches midpoint",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  10,
				ScoringType: ScoringDynamic,
			},
			wantPoints: 100,
		},
		{
			name: "more than decay solves stays at min",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  20,
				ScoringType: ScoringDynamic,
			},
			wantPoints: 100,
		},
		{
			name: "zero decay defaults to 10",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       0,
				SolveCount:  5,
				ScoringType: ScoringDynamic,
			},
			wantPoints: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.challenge.CurrentPoints()
			if got != tt.wantPoints {
				t.Errorf("CurrentPoints() = %d, want %d", got, tt.wantPoints)
			}
		})
	}
}

func TestCurrentPoints_StaticScoring(t *testing.T) {
	tests := []struct {
		name       string
		challenge  Challenge
		wantPoints int
	}{
		{
			name: "static always returns max points with 0 solves",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  0,
				ScoringType: ScoringStatic,
			},
			wantPoints: 500,
		},
		{
			name: "static always returns max points with many solves",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  100,
				ScoringType: ScoringStatic,
			},
			wantPoints: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.challenge.CurrentPoints()
			if got != tt.wantPoints {
				t.Errorf("CurrentPoints() = %d, want %d", got, tt.wantPoints)
			}
		})
	}
}

func TestCurrentPoints_LinearScoring(t *testing.T) {
	tests := []struct {
		name       string
		challenge  Challenge
		wantPoints int
	}{
		{
			name: "linear no solves returns max",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  0,
				ScoringType: ScoringLinear,
			},
			wantPoints: 500,
		},
		{
			name: "linear mid solves decreases linearly",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  5,
				ScoringType: ScoringLinear,
			},
			wantPoints: 300,
		},
		{
			name: "linear at decay solves reaches min",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  10,
				ScoringType: ScoringLinear,
			},
			wantPoints: 100,
		},
		{
			name: "linear beyond decay stays at min",
			challenge: Challenge{
				MaxPoints:   500,
				MinPoints:   100,
				Decay:       10,
				SolveCount:  20,
				ScoringType: ScoringLinear,
			},
			wantPoints: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.challenge.CurrentPoints()
			if got != tt.wantPoints {
				t.Errorf("CurrentPoints() = %d, want %d", got, tt.wantPoints)
			}
		})
	}
}

func TestCurrentPoints_DefaultScoringType(t *testing.T) {
	// Empty scoring type should default to dynamic
	c := Challenge{
		MaxPoints:  500,
		MinPoints:  100,
		Decay:      10,
		SolveCount: 0,
	}
	if got := c.CurrentPoints(); got != 500 {
		t.Errorf("CurrentPoints() with empty scoring_type = %d, want 500", got)
	}
}
