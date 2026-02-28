package models

type Hint struct {
	ID          string `json:"id"`
	ChallengeID string `json:"challenge_id"`
	Content     string `json:"content"`
	Cost        int    `json:"cost"`
	Order       int    `json:"order"`
}

type HintReveal struct {
	ID          string `json:"id"`
	HintID      string `json:"hint_id"`
	ChallengeID string `json:"challenge_id"`
	UserID      string `json:"user_id"`
	TeamID      string `json:"team_id,omitempty"`
	Cost        int    `json:"cost"`
}
