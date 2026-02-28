package repositories

import (
	"database/sql"
)

type ContestSolveRepository struct {
	db *sql.DB
}

func NewContestSolveRepository(db *sql.DB) *ContestSolveRepository {
	return &ContestSolveRepository{db: db}
}

// GetContestSolveCount returns the solve count for a challenge in a specific contest.
func (r *ContestSolveRepository) GetContestSolveCount(contestID, challengeID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT solve_count FROM contest_challenge_solves WHERE contest_id=? AND challenge_id=?",
		contestID, challengeID,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// IncrementContestSolveCount atomically increments the solve count for a challenge in a contest.
// Creates the row if it doesn't exist (upsert).
func (r *ContestSolveRepository) IncrementContestSolveCount(contestID, challengeID string) error {
	query := `INSERT INTO contest_challenge_solves (contest_id, challenge_id, solve_count) VALUES (?, ?, 1)
			  ON CONFLICT(contest_id, challenge_id) DO UPDATE SET solve_count = solve_count + 1`
	_, err := r.db.Exec(query, contestID, challengeID)
	return err
}

// GetContestSolveCounts returns solve counts for all challenges in a contest.
func (r *ContestSolveRepository) GetContestSolveCounts(contestID string) (map[string]int, error) {
	rows, err := r.db.Query(
		"SELECT challenge_id, solve_count FROM contest_challenge_solves WHERE contest_id=?",
		contestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var challengeID string
		var count int
		if err := rows.Scan(&challengeID, &count); err != nil {
			return nil, err
		}
		result[challengeID] = count
	}
	return result, nil
}
