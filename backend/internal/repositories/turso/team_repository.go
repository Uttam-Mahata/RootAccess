package turso

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const teamColumns = `id, name, description, avatar, leader_id, invite_code, score, created_at, updated_at`

// validTeamColumns guards against SQL injection in dynamic UpdateTeamFields queries.
var validTeamColumns = map[string]bool{
	"name":        true,
	"description": true,
	"avatar":      true,
	"invite_code": true,
	"leader_id":   true,
}

// TeamRepository implements interfaces.TeamRepository using database/sql.
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new Turso-backed TeamRepository.
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// scanTeam scans a single team row (without member IDs).
func scanTeam(s scanner) (*models.Team, error) {
	var t models.Team
	var (
		id, leaderID           string
		name, desc, avatar     sql.NullString
		inviteCode             sql.NullString
		score                  int
		createdAt, updatedAt   sql.NullString
	)

	err := s.Scan(&id, &name, &desc, &avatar, &leaderID, &inviteCode, &score, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	t.ID = oidFromHex(id)
	t.Name = nullStr(name)
	t.Description = nullStr(desc)
	t.Avatar = nullStr(avatar)
	t.LeaderID = oidFromHex(leaderID)
	t.InviteCode = nullStr(inviteCode)
	t.Score = score
	t.CreatedAt = strToTime(nullStr(createdAt))
	t.UpdatedAt = strToTime(nullStr(updatedAt))

	return &t, nil
}

// loadTeamMembers loads member IDs from the join table.
func (r *TeamRepository) loadTeamMembers(teamID string) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(`SELECT member_id FROM team_member_ids WHERE team_id = ?`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []primitive.ObjectID
	for rows.Next() {
		var mid string
		if err := rows.Scan(&mid); err != nil {
			return nil, err
		}
		members = append(members, oidFromHex(mid))
	}
	return members, rows.Err()
}

// findOneTeam queries a single team and loads its members.
func (r *TeamRepository) findOneTeam(where string, args ...interface{}) (*models.Team, error) {
	row := r.db.QueryRow("SELECT "+teamColumns+" FROM teams WHERE "+where, args...)
	t, err := scanTeam(row)
	if err != nil {
		return nil, err
	}
	t.MemberIDs, err = r.loadTeamMembers(t.ID.Hex())
	if err != nil {
		return nil, err
	}
	return t, nil
}

// queryTeams queries multiple teams and loads their members.
func (r *TeamRepository) queryTeams(query string, args ...interface{}) ([]models.Team, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		t.MemberIDs, err = r.loadTeamMembers(t.ID.Hex())
		if err != nil {
			return nil, err
		}
		teams = append(teams, *t)
	}
	return teams, rows.Err()
}

// ── Interface methods ──────────────────────────────────────────────────

func (r *TeamRepository) CreateTeam(team *models.Team) error {
	now := time.Now()
	id := newID()
	team.CreatedAt = now
	team.UpdatedAt = now

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO teams (`+teamColumns+`) VALUES (?,?,?,?,?,?,?,?,?)`,
		id, team.Name, team.Description, team.Avatar,
		team.LeaderID.Hex(), team.InviteCode, team.Score,
		timeToStr(now), timeToStr(now),
	)
	if err != nil {
		return err
	}

	for _, mid := range team.MemberIDs {
		_, err = tx.Exec(`INSERT INTO team_member_ids (team_id, member_id) VALUES (?, ?)`, id, mid.Hex())
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	team.ID = oidFromHex(id)
	return nil
}

func (r *TeamRepository) FindTeamByID(teamID string) (*models.Team, error) {
	return r.findOneTeam("id = ?", teamID)
}

func (r *TeamRepository) FindTeamByLeaderID(leaderID string) (*models.Team, error) {
	return r.findOneTeam("leader_id = ?", leaderID)
}

func (r *TeamRepository) FindTeamByMemberID(userID string) (*models.Team, error) {
	row := r.db.QueryRow(
		`SELECT `+teamColumns+` FROM teams
		 INNER JOIN team_member_ids ON teams.id = team_member_ids.team_id
		 WHERE team_member_ids.member_id = ?`, userID,
	)
	t, err := scanTeam(row)
	if err != nil {
		return nil, err
	}
	t.MemberIDs, err = r.loadTeamMembers(t.ID.Hex())
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TeamRepository) FindTeamByInviteCode(code string) (*models.Team, error) {
	return r.findOneTeam("invite_code = ?", code)
}

func (r *TeamRepository) FindTeamByName(name string) (*models.Team, error) {
	return r.findOneTeam("name = ?", name)
}

func (r *TeamRepository) UpdateTeam(team *models.Team) error {
	team.UpdatedAt = time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE teams SET name=?, description=?, avatar=?, leader_id=?, invite_code=?, score=?, updated_at=? WHERE id=?`,
		team.Name, team.Description, team.Avatar,
		team.LeaderID.Hex(), team.InviteCode, team.Score,
		timeToStr(team.UpdatedAt), team.ID.Hex(),
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM team_member_ids WHERE team_id = ?`, team.ID.Hex())
	if err != nil {
		return err
	}

	for _, mid := range team.MemberIDs {
		_, err = tx.Exec(`INSERT INTO team_member_ids (team_id, member_id) VALUES (?, ?)`, team.ID.Hex(), mid.Hex())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TeamRepository) DeleteTeam(teamID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM team_member_ids WHERE team_id = ?`, teamID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM teams WHERE id = ?`, teamID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TeamRepository) AddMemberToTeam(teamID, userID string) error {
	_, err := r.db.Exec(`INSERT INTO team_member_ids (team_id, member_id) VALUES (?, ?)`, teamID, userID)
	return err
}

func (r *TeamRepository) RemoveMemberFromTeam(teamID, userID string) error {
	_, err := r.db.Exec(`DELETE FROM team_member_ids WHERE team_id = ? AND member_id = ?`, teamID, userID)
	return err
}

func (r *TeamRepository) UpdateTeamScore(teamID string, points int) error {
	_, err := r.db.Exec(`UPDATE teams SET score = score + ?, updated_at = ? WHERE id = ?`,
		points, timeToStr(time.Now()), teamID)
	return err
}

func (r *TeamRepository) GetAllTeamsWithScores() ([]models.Team, error) {
	return r.queryTeams("SELECT " + teamColumns + " FROM teams ORDER BY score DESC")
}

func (r *TeamRepository) CountTeams() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM teams`).Scan(&count)
	return count, err
}

func (r *TeamRepository) GetTeamMemberCount(teamID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM team_member_ids WHERE team_id = ?`, teamID).Scan(&count)
	return count, err
}

func (r *TeamRepository) GetAllTeams() ([]models.Team, error) {
	return r.queryTeams("SELECT " + teamColumns + " FROM teams ORDER BY created_at DESC")
}

func (r *TeamRepository) UpdateTeamFields(teamID primitive.ObjectID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	var setClauses []string
	var args []interface{}

	for key, val := range fields {
		if !validTeamColumns[key] {
			continue
		}
		switch v := val.(type) {
		case primitive.ObjectID:
			setClauses = append(setClauses, key+" = ?")
			args = append(args, v.Hex())
		default:
			setClauses = append(setClauses, key+" = ?")
			args = append(args, val)
		}
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, timeToStr(time.Now()))

	if len(setClauses) == 1 {
		// Only updated_at — nothing valid to update
		return nil
	}

	args = append(args, teamID.Hex())
	query := "UPDATE teams SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *TeamRepository) AdminDeleteTeam(teamID primitive.ObjectID) error {
	return r.DeleteTeam(teamID.Hex())
}

func (r *TeamRepository) AdminUpdateTeamLeader(teamID, newLeaderID primitive.ObjectID) error {
	_, err := r.db.Exec(`UPDATE teams SET leader_id = ?, updated_at = ? WHERE id = ?`,
		newLeaderID.Hex(), timeToStr(time.Now()), teamID.Hex())
	return err
}

func (r *TeamRepository) GetRecentTeams(since time.Time) ([]models.Team, error) {
	return r.queryTeams(
		"SELECT "+teamColumns+" FROM teams WHERE created_at >= ? ORDER BY created_at DESC",
		timeToStr(since),
	)
}

