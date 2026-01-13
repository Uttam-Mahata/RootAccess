package models

import (
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	SolveCount  int                `bson:"solve_count" json:"solve_count"`
	FlagHash    string             `bson:"flag_hash" json:"-"` // SHA-256 hashed flag (hidden from API)
	Files       []string           `bson:"files" json:"files"`
}

// CurrentPoints calculates dynamic points based on solve count using CTFd formula
// Formula: value = ((min - max) / decay^2) * solves^2 + max
// This creates a smooth curve where points decrease as more teams solve
func (c *Challenge) CurrentPoints() int {
	if c.SolveCount <= 0 {
		return c.MaxPoints
	}

	if c.Decay <= 0 {
		c.Decay = 10 // Default decay
	}

	// CTFd standard formula
	decaySquared := float64(c.Decay * c.Decay)
	solvesSquared := float64(c.SolveCount * c.SolveCount)
	
	value := ((float64(c.MinPoints) - float64(c.MaxPoints)) / decaySquared) * solvesSquared + float64(c.MaxPoints)
	
	// Ensure points don't go below minimum
	points := int(math.Round(value))
	if points < c.MinPoints {
		return c.MinPoints
	}
	
	return points
}
