package models

import (
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Scoring type constants
const (
	ScoringStatic  = "static"  // Fixed points, never changes
	ScoringLinear  = "linear"  // Points decrease linearly with solves
	ScoringDynamic = "dynamic" // Points decrease quadratically (CTFd formula)
)

type Challenge struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	Difficulty  string             `bson:"difficulty" json:"difficulty"` // easy, medium, hard
	MaxPoints   int                `bson:"max_points" json:"max_points"` // Maximum/initial points
	MinPoints   int                `bson:"min_points" json:"min_points"` // Minimum floor points
	Decay       int                `bson:"decay" json:"decay"`           // Decay factor (solves to reach midpoint)
	ScoringType string             `bson:"scoring_type" json:"scoring_type"` // static, linear, dynamic
	SolveCount  int                `bson:"solve_count" json:"solve_count"`
	FlagHash    string             `bson:"flag_hash" json:"-"` // SHA-256 hashed flag (hidden from API)
	Files       []string           `bson:"files" json:"files"`
	Hints       []Hint             `bson:"hints,omitempty" json:"hints,omitempty"` // Embedded hints
}

// CurrentPoints calculates points based on scoring type and solve count
func (c *Challenge) CurrentPoints() int {
	switch c.ScoringType {
	case ScoringStatic:
		return c.MaxPoints
	case ScoringLinear:
		return c.linearPoints()
	case ScoringDynamic:
		return c.dynamicPoints()
	default:
		// Default to dynamic for backward compatibility
		return c.dynamicPoints()
	}
}

// dynamicPoints calculates points using CTFd formula
// Formula: value = ((min - max) / decay^2) * solves^2 + max
func (c *Challenge) dynamicPoints() int {
	if c.SolveCount <= 0 {
		return c.MaxPoints
	}

	if c.Decay <= 0 {
		c.Decay = 10 // Default decay
	}

	decaySquared := float64(c.Decay * c.Decay)
	solvesSquared := float64(c.SolveCount * c.SolveCount)

	value := ((float64(c.MinPoints) - float64(c.MaxPoints)) / decaySquared) * solvesSquared + float64(c.MaxPoints)

	points := int(math.Round(value))
	if points < c.MinPoints {
		return c.MinPoints
	}

	return points
}

// linearPoints calculates points using a linear decay formula
// Points decrease linearly from MaxPoints to MinPoints over Decay solves
func (c *Challenge) linearPoints() int {
	if c.SolveCount <= 0 {
		return c.MaxPoints
	}

	if c.Decay <= 0 {
		c.Decay = 10
	}

	decrease := float64(c.MaxPoints-c.MinPoints) * float64(c.SolveCount) / float64(c.Decay)
	points := int(math.Round(float64(c.MaxPoints) - decrease))

	if points < c.MinPoints {
		return c.MinPoints
	}

	return points
}
