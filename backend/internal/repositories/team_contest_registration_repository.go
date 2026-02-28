package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TeamContestRegistrationRepository struct {
	db *sql.DB
}

func NewTeamContestRegistrationRepository(db *sql.DB) *TeamContestRegistrationRepository {
	return &TeamContestRegistrationRepository{db: db}
}

func (r *TeamContestRegistrationRepository) RegisterTeam(teamID, contestID string) error {
	id := uuid.New().String()
	_, err := r.db.Exec("INSERT OR IGNORE INTO team_contest_registrations (id, team_id, contest_id, registered_at) VALUES (?, ?, ?, ?)",
		id, teamID, contestID, time.Now().Format(time.RFC3339))
	return err
}

func (r *TeamContestRegistrationRepository) UnregisterTeam(teamID, contestID string) error {
	_, err := r.db.Exec("DELETE FROM team_contest_registrations WHERE team_id=? AND contest_id=?", teamID, contestID)
	return err
}

func (r *TeamContestRegistrationRepository) IsTeamRegistered(teamID, contestID string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM team_contest_registrations WHERE team_id=? AND contest_id=?", teamID, contestID).Scan(&count)
	return count > 0, err
}

func (r *TeamContestRegistrationRepository) GetTeamContests(teamID string) ([]string, error) {
	rows, err := r.db.Query("SELECT contest_id FROM team_contest_registrations WHERE team_id=?", teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *TeamContestRegistrationRepository) CountContestTeams(contestID string) (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM team_contest_registrations WHERE contest_id=?", contestID).Scan(&count)
	return count, err
}

func (r *TeamContestRegistrationRepository) GetContestTeams(contestID string) ([]string, error) {
	rows, err := r.db.Query("SELECT team_id FROM team_contest_registrations WHERE contest_id=?", contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
