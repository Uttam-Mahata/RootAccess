package turso

import (
	"database/sql"
	"fmt"
	"log"
)

// RunMigrations creates all tables required by the RootAccess CTF platform.
// All IDs are stored as TEXT (hex ObjectID format) and timestamps as RFC 3339 TEXT strings.
func RunMigrations(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range migrationStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration: %w\nstatement: %s", err, stmt)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration transaction: %w", err)
	}

	log.Println("turso: all migrations completed successfully")
	return nil
}

var migrationStatements = []string{
	// ── users ───────────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS users (
		id                  TEXT PRIMARY KEY,
		username            TEXT UNIQUE NOT NULL,
		email               TEXT UNIQUE NOT NULL,
		password_hash       TEXT,
		role                TEXT DEFAULT 'user',
		email_verified      INTEGER DEFAULT 0,
		verification_token  TEXT,
		verification_expiry TEXT,
		reset_password_token  TEXT,
		reset_password_expiry TEXT,
		oauth_provider      TEXT,
		oauth_provider_id   TEXT,
		oauth_access_token  TEXT,
		oauth_refresh_token TEXT,
		oauth_expires_at    TEXT,
		status              TEXT DEFAULT 'active',
		ban_reason          TEXT,
		suspended_until     TEXT,
		last_ip             TEXT,
		last_login_at       TEXT,
		created_at          TEXT,
		updated_at          TEXT
	)`,

	// ── user_ip_history ─────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS user_ip_history (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id   TEXT REFERENCES users(id),
		ip        TEXT,
		action    TEXT,
		timestamp TEXT
	)`,

	// ── challenges ──────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS challenges (
		id                         TEXT PRIMARY KEY,
		title                      TEXT,
		description                TEXT,
		description_format         TEXT DEFAULT 'markdown',
		category                   TEXT,
		difficulty                 TEXT,
		max_points                 INTEGER,
		min_points                 INTEGER,
		decay                      INTEGER DEFAULT 10,
		scoring_type               TEXT DEFAULT 'dynamic',
		solve_count                INTEGER DEFAULT 0,
		flag_hash                  TEXT,
		tags                       TEXT,
		files                      TEXT,
		scheduled_at               TEXT,
		is_published               INTEGER DEFAULT 0,
		contest_id                 TEXT,
		official_writeup           TEXT,
		official_writeup_format    TEXT,
		official_writeup_published INTEGER DEFAULT 0
	)`,

	// ── submissions ─────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS submissions (
		id           TEXT PRIMARY KEY,
		user_id      TEXT,
		team_id      TEXT,
		challenge_id TEXT,
		flag         TEXT,
		is_correct   INTEGER DEFAULT 0,
		ip_address   TEXT,
		timestamp    TEXT
	)`,

	// ── teams ───────────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS teams (
		id          TEXT PRIMARY KEY,
		name        TEXT UNIQUE NOT NULL,
		description TEXT,
		avatar      TEXT,
		leader_id   TEXT,
		invite_code TEXT UNIQUE,
		score       INTEGER DEFAULT 0,
		created_at  TEXT,
		updated_at  TEXT
	)`,

	// ── team_member_ids ─────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS team_member_ids (
		team_id   TEXT,
		member_id TEXT,
		PRIMARY KEY (team_id, member_id)
	)`,

	// ── team_invitations ────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS team_invitations (
		id            TEXT PRIMARY KEY,
		team_id       TEXT,
		team_name     TEXT,
		inviter_id    TEXT,
		inviter_name  TEXT,
		invitee_email TEXT,
		invitee_user_id TEXT,
		token         TEXT UNIQUE,
		status        TEXT DEFAULT 'pending',
		expires_at    TEXT,
		created_at    TEXT
	)`,

	// ── notifications ───────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS notifications (
		id         TEXT PRIMARY KEY,
		title      TEXT,
		content    TEXT,
		type       TEXT DEFAULT 'info',
		created_by TEXT,
		created_at TEXT,
		is_active  INTEGER DEFAULT 1
	)`,

	// ── hint_reveals ────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS hint_reveals (
		id           TEXT PRIMARY KEY,
		hint_id      TEXT,
		challenge_id TEXT,
		user_id      TEXT,
		team_id      TEXT,
		cost         INTEGER
	)`,

	// ── contest_config ──────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS contest_config (
		id                    TEXT PRIMARY KEY,
		contest_id            TEXT,
		start_time            TEXT,
		end_time              TEXT,
		freeze_time           TEXT,
		title                 TEXT,
		is_active             INTEGER DEFAULT 0,
		is_paused             INTEGER DEFAULT 0,
		scoreboard_visibility TEXT DEFAULT 'public',
		updated_at            TEXT
	)`,

	// ── contests ────────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS contests (
		id                    TEXT PRIMARY KEY,
		name                  TEXT,
		description           TEXT,
		start_time            TEXT,
		end_time              TEXT,
		freeze_time           TEXT,
		scoreboard_visibility TEXT DEFAULT 'public',
		is_active             INTEGER DEFAULT 0,
		created_at            TEXT,
		updated_at            TEXT
	)`,

	// ── contest_rounds ──────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS contest_rounds (
		id           TEXT PRIMARY KEY,
		contest_id   TEXT,
		name         TEXT,
		description  TEXT,
		"order"      INTEGER,
		visible_from TEXT,
		start_time   TEXT,
		end_time     TEXT,
		created_at   TEXT,
		updated_at   TEXT
	)`,

	// ── round_challenges ────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS round_challenges (
		id           TEXT PRIMARY KEY,
		round_id     TEXT,
		challenge_id TEXT,
		created_at   TEXT,
		UNIQUE(round_id, challenge_id)
	)`,

	// ── writeups ────────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS writeups (
		id             TEXT PRIMARY KEY,
		challenge_id   TEXT,
		user_id        TEXT,
		username       TEXT,
		content        TEXT,
		content_format TEXT DEFAULT 'markdown',
		status         TEXT DEFAULT 'pending',
		upvotes        INTEGER DEFAULT 0,
		created_at     TEXT,
		updated_at     TEXT
	)`,

	// ── writeup_upvotes ─────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS writeup_upvotes (
		writeup_id TEXT,
		user_id    TEXT,
		PRIMARY KEY (writeup_id, user_id)
	)`,

	// ── audit_logs ──────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS audit_logs (
		id         TEXT PRIMARY KEY,
		user_id    TEXT,
		username   TEXT,
		action     TEXT,
		resource   TEXT,
		details    TEXT,
		ip_address TEXT,
		created_at TEXT
	)`,

	// ── achievements ────────────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS achievements (
		id           TEXT PRIMARY KEY,
		user_id      TEXT,
		team_id      TEXT,
		type         TEXT,
		name         TEXT,
		description  TEXT,
		icon         TEXT,
		challenge_id TEXT,
		category     TEXT,
		earned_at    TEXT
	)`,

	// ── score_adjustments ───────────────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS score_adjustments (
		id          TEXT PRIMARY KEY,
		target_type TEXT,
		target_id   TEXT,
		delta       INTEGER,
		reason      TEXT,
		created_by  TEXT,
		created_at  TEXT
	)`,

	// ── team_contest_registrations ──────────────────────────────────────
	`CREATE TABLE IF NOT EXISTS team_contest_registrations (
		id            TEXT PRIMARY KEY,
		team_id       TEXT,
		contest_id    TEXT,
		registered_at TEXT,
		UNIQUE(team_id, contest_id)
	)`,
}
