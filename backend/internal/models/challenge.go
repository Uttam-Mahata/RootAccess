package models

import (
	"math"
)

const (
	ScoringStatic  = "static"
	ScoringLinear  = "linear"
	ScoringDynamic = "dynamic"
)

type Challenge struct {
	ID                       string   `json:"id"`
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	DescriptionFormat        string   `json:"description_format"`
	Category                 string   `json:"category"`
	Difficulty               string   `json:"difficulty"`
	MaxPoints                int      `json:"max_points"`
	MinPoints                int      `json:"min_points"`
	Decay                    int      `json:"decay"`
	ScoringType              string   `json:"scoring_type"`
	SolveCount               int      `json:"solve_count"`
	FlagHash                 string   `json:"-"`
	Files                    []string `json:"files"`
	Tags                     []string `json:"tags"`
	ScheduledAt              string   `json:"scheduled_at,omitempty"`
	IsPublished              bool     `json:"is_published"`
	Hints                    []Hint   `json:"hints,omitempty"`
	ContestID                string   `json:"contest_id,omitempty"`
	OfficialWriteup          string   `json:"official_writeup,omitempty"`
	OfficialWriteupFormat    string   `json:"official_writeup_format,omitempty"`
	OfficialWriteupPublished bool     `json:"official_writeup_published"`
}

func (c *Challenge) CurrentPoints() int {
	switch c.ScoringType {
	case ScoringStatic:
		return c.MaxPoints
	case ScoringLinear:
		return c.linearPoints()
	case ScoringDynamic:
		return c.dynamicPoints()
	default:
		return c.dynamicPoints()
	}
}

func (c *Challenge) dynamicPoints() int {
	if c.SolveCount <= 0 {
		return c.MaxPoints
	}

	decay := c.Decay
	if decay <= 0 {
		decay = 10
	}

	decaySquared := float64(decay * decay)
	solvesSquared := float64(c.SolveCount * c.SolveCount)

	value := ((float64(c.MinPoints)-float64(c.MaxPoints))/decaySquared)*solvesSquared + float64(c.MaxPoints)

	points := int(math.Round(value))
	if points < c.MinPoints {
		return c.MinPoints
	}

	return points
}

func (c *Challenge) linearPoints() int {
	if c.SolveCount <= 0 {
		return c.MaxPoints
	}

	decay := c.Decay
	if decay <= 0 {
		decay = 10
	}

	decrease := float64(c.MaxPoints-c.MinPoints) * float64(c.SolveCount) / float64(decay)
	points := int(math.Round(float64(c.MaxPoints) - decrease))

	if points < c.MinPoints {
		return c.MinPoints
	}

	return points
}
