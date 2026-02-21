package turso

import (
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamContestRegistrationRepository implements interfaces.TeamContestRegistrationRepository using database/sql.
type TeamContestRegistrationRepository struct {
	db *sql.DB
}

// NewTeamContestRegistrationRepository creates a new Turso-backed TeamContestRegistrationRepository.
func NewTeamContestRegistrationRepository(db *sql.DB) *TeamContestRegistrationRepository {
	return &TeamContestRegistrationRepository{db: db}
}

func (r *TeamContestRegistrationRepository) CreateIndexes() error {
	// No-op: SQLite handles UNIQUE constraint via schema definition.
	return nil
}

func (r *TeamContestRegistrationRepository) RegisterTeam(teamID, contestID primitive.ObjectID) error {
	id := newID()
	_, err := r.db.Exec(
		`INSERT INTO team_contest_registrations (id, team_id, contest_id, registered_at) VALUES (?,?,?,?)`,
		id, teamID.Hex(), contestID.Hex(), timeToStr(time.Now()),
	)
	return err
}

func (r *TeamContestRegistrationRepository) UnregisterTeam(teamID, contestID primitive.ObjectID) error {
	_, err := r.db.Exec(
		`DELETE FROM team_contest_registrations WHERE team_id = ? AND contest_id = ?`,
		teamID.Hex(), contestID.Hex(),
	)
	return err
}

func (r *TeamContestRegistrationRepository) IsTeamRegistered(teamID, contestID primitive.ObjectID) (bool, error) {
	var exists int
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM team_contest_registrations WHERE team_id = ? AND contest_id = ?)`,
		teamID.Hex(), contestID.Hex(),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *TeamContestRegistrationRepository) GetTeamContests(teamID primitive.ObjectID) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(
		`SELECT contest_id FROM team_contest_registrations WHERE team_id = ?`,
		teamID.Hex(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(cid))
	}
	return ids, rows.Err()
}

func (r *TeamContestRegistrationRepository) CountContestTeams(contestID primitive.ObjectID) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM team_contest_registrations WHERE contest_id = ?`,
		contestID.Hex(),
	).Scan(&count)
	return count, err
}

func (r *TeamContestRegistrationRepository) GetContestTeams(contestID primitive.ObjectID) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(
		`SELECT team_id FROM team_contest_registrations WHERE contest_id = ?`,
		contestID.Hex(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(tid))
	}
	return ids, rows.Err()
}
