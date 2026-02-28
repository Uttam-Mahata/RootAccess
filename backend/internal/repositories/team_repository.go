package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(team *models.Team) error {
	if team.ID == "" {
		team.ID = uuid.New().String()
	}
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO teams (id, name, description, avatar, leader_id, invite_code, score, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(query, team.ID, team.Name, team.Description, team.Avatar, team.LeaderID, team.InviteCode, team.Score, team.CreatedAt.Format(time.RFC3339), team.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO team_members (team_id, user_id) VALUES (?, ?)", team.ID, team.LeaderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TeamRepository) scanTeam(row *sql.Row) (*models.Team, error) {
	var team models.Team
	var createdAt, updatedAt string
	err := row.Scan(
		&team.ID, &team.Name, &team.Description, &team.Avatar,
		&team.LeaderID, &team.InviteCode, &team.Score,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}
	team.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	team.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &team, nil
}

func (r *TeamRepository) selectTeamFields() string {
	return "id, name, description, avatar, leader_id, invite_code, score, created_at, updated_at"
}

func (r *TeamRepository) FindTeamByID(teamID string) (*models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams WHERE id=?", r.selectTeamFields())
	return r.scanTeam(r.db.QueryRow(query, teamID))
}

func (r *TeamRepository) FindTeamByLeaderID(leaderID string) (*models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams WHERE leader_id=?", r.selectTeamFields())
	return r.scanTeam(r.db.QueryRow(query, leaderID))
}

func (r *TeamRepository) FindTeamByMemberID(userID string) (*models.Team, error) {
	query := fmt.Sprintf(`
		SELECT t.id, t.name, t.description, t.avatar, t.leader_id, t.invite_code, t.score, t.created_at, t.updated_at 
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = ?
	`)
	return r.scanTeam(r.db.QueryRow(query, userID))
}

func (r *TeamRepository) FindTeamByInviteCode(code string) (*models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams WHERE invite_code=?", r.selectTeamFields())
	return r.scanTeam(r.db.QueryRow(query, code))
}

func (r *TeamRepository) FindTeamByName(name string) (*models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams WHERE name=?", r.selectTeamFields())
	return r.scanTeam(r.db.QueryRow(query, name))
}

func (r *TeamRepository) UpdateTeam(team *models.Team) error {
	team.UpdatedAt = time.Now()
	query := `UPDATE teams SET name=?, description=?, avatar=?, leader_id=?, invite_code=?, score=?, updated_at=? WHERE id=?`
	_, err := r.db.Exec(query, team.Name, team.Description, team.Avatar, team.LeaderID, team.InviteCode, team.Score, team.UpdatedAt.Format(time.RFC3339), team.ID)
	return err
}

func (r *TeamRepository) DeleteTeam(teamID string) error {
	_, err := r.db.Exec("DELETE FROM teams WHERE id=?", teamID)
	return err
}

func (r *TeamRepository) AddMemberToTeam(teamID, userID string) error {
	_, err := r.db.Exec("INSERT OR IGNORE INTO team_members (team_id, user_id) VALUES (?, ?)", teamID, userID)
	if err == nil {
		_, err = r.db.Exec("UPDATE teams SET updated_at=? WHERE id=?", time.Now().Format(time.RFC3339), teamID)
	}
	return err
}

func (r *TeamRepository) RemoveMemberFromTeam(teamID, userID string) error {
	_, err := r.db.Exec("DELETE FROM team_members WHERE team_id=? AND user_id=?", teamID, userID)
	if err == nil {
		_, err = r.db.Exec("UPDATE teams SET updated_at=? WHERE id=?", time.Now().Format(time.RFC3339), teamID)
	}
	return err
}

func (r *TeamRepository) UpdateTeamScore(teamID string, points int) error {
	query := `UPDATE teams SET score = score + ?, updated_at = ? WHERE id=?`
	_, err := r.db.Exec(query, points, time.Now().Format(time.RFC3339), teamID)
	return err
}

func (r *TeamRepository) GetAllTeamsWithScores() ([]models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams ORDER BY score DESC", r.selectTeamFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		var createdAt, updatedAt string
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.Avatar,
			&team.LeaderID, &team.InviteCode, &team.Score,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		team.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		team.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *TeamRepository) CountTeams() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM teams").Scan(&count)
	return count, err
}

func (r *TeamRepository) GetTeamMemberCount(teamID string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM team_members WHERE team_id=?", teamID).Scan(&count)
	return count, err
}

func (r *TeamRepository) GetAllTeams() ([]models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams ORDER BY created_at DESC", r.selectTeamFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		var createdAt, updatedAt string
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.Avatar,
			&team.LeaderID, &team.InviteCode, &team.Score,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		team.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		team.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *TeamRepository) UpdateTeamFields(teamID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now().Format(time.RFC3339)

	var setClauses []string
	var args []interface{}

	for k, v := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s=?", k))
		args = append(args, v)
	}
	args = append(args, teamID)

	query := fmt.Sprintf("UPDATE teams SET %s WHERE id=?", strings.Join(setClauses, ", "))
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *TeamRepository) AdminDeleteTeam(teamID string) error {
	_, err := r.db.Exec("DELETE FROM teams WHERE id=?", teamID)
	return err
}

func (r *TeamRepository) AdminUpdateTeamLeader(teamID, newLeaderID string) error {
	update := `UPDATE teams SET leader_id=?, updated_at=? WHERE id=?`
	_, err := r.db.Exec(update, newLeaderID, time.Now().Format(time.RFC3339), teamID)
	return err
}

func (r *TeamRepository) GetRecentTeams(since time.Time) ([]models.Team, error) {
	query := fmt.Sprintf("SELECT %s FROM teams WHERE created_at >= ? ORDER BY created_at DESC", r.selectTeamFields())
	rows, err := r.db.Query(query, since.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		var createdAt, updatedAt string
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.Avatar,
			&team.LeaderID, &team.InviteCode, &team.Score,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		team.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		team.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *TeamRepository) GetTeamMembers(teamID string) ([]string, error) {
	rows, err := r.db.Query("SELECT user_id FROM team_members WHERE team_id=?", teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var member string
		if err := rows.Scan(&member); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}
