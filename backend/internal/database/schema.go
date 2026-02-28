package database

import (
	"database/sql"
	"log"
)

func BootstrapSchema(db *sql.DB) {
	schemas := []string{
		// Users
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			email_verified INTEGER NOT NULL DEFAULT 0,
			verification_token TEXT,
			verification_expiry TEXT,
			reset_password_token TEXT,
			reset_password_expiry TEXT,
			oauth_provider TEXT,
			oauth_provider_id TEXT,
			oauth_access_token TEXT,
			oauth_refresh_token TEXT,
			oauth_expires_at TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			ban_reason TEXT,
			suspended_until TEXT,
			last_ip TEXT,
			last_login_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Teams
		`CREATE TABLE IF NOT EXISTS teams (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			avatar TEXT,
			leader_id TEXT NOT NULL REFERENCES users(id),
			invite_code TEXT UNIQUE NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Team Members (Junction)
		`CREATE TABLE IF NOT EXISTS team_members (
			team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY (team_id, user_id)
		);`,
		// Team Invitations
		`CREATE TABLE IF NOT EXISTS team_invitations (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			team_name TEXT NOT NULL,
			inviter_id TEXT NOT NULL REFERENCES users(id),
			inviter_name TEXT NOT NULL,
			invitee_email TEXT,
			invitee_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
			token TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		// Challenges
		`CREATE TABLE IF NOT EXISTS challenges (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			description_format TEXT NOT NULL,
			category TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			max_points INTEGER NOT NULL,
			min_points INTEGER NOT NULL,
			decay INTEGER NOT NULL,
			scoring_type TEXT NOT NULL,
			solve_count INTEGER NOT NULL DEFAULT 0,
			flag_hash TEXT NOT NULL,
			files TEXT,
			tags TEXT,
			scheduled_at TEXT,
			is_published INTEGER NOT NULL DEFAULT 0,
			contest_id TEXT REFERENCES contests(id) ON DELETE SET NULL,
			official_writeup TEXT,
			official_writeup_format TEXT,
			official_writeup_published INTEGER NOT NULL DEFAULT 0
		);`,
		// Hints
		`CREATE TABLE IF NOT EXISTS hints (
			id TEXT PRIMARY KEY,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			cost INTEGER NOT NULL,
			display_order INTEGER NOT NULL
		);`,
		// Hint Reveals
		`CREATE TABLE IF NOT EXISTS hint_reveals (
			id TEXT PRIMARY KEY,
			hint_id TEXT NOT NULL REFERENCES hints(id) ON DELETE CASCADE,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
			cost INTEGER NOT NULL
		);`,
		// Submissions
		`CREATE TABLE IF NOT EXISTS submissions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			contest_id TEXT REFERENCES contests(id) ON DELETE SET NULL,
			flag TEXT NOT NULL,
			is_correct INTEGER NOT NULL,
			ip_address TEXT,
			timestamp TEXT NOT NULL
		);`,
		// Notifications
		`CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL,
			created_by TEXT NOT NULL REFERENCES users(id),
			created_at TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);`,
		// Writeups
		`CREATE TABLE IF NOT EXISTS writeups (
			id TEXT PRIMARY KEY,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			username TEXT NOT NULL,
			content TEXT NOT NULL,
			content_format TEXT NOT NULL,
			status TEXT NOT NULL,
			upvotes INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Writeup Upvotes (Junction)
		`CREATE TABLE IF NOT EXISTS writeup_upvotes (
			writeup_id TEXT NOT NULL REFERENCES writeups(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY (writeup_id, user_id)
		);`,
		// Achievements
		`CREATE TABLE IF NOT EXISTS achievements (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			icon TEXT NOT NULL,
			challenge_id TEXT REFERENCES challenges(id) ON DELETE SET NULL,
			category TEXT,
			earned_at TEXT NOT NULL
		);`,
		// Audit Logs
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			username TEXT NOT NULL,
			action TEXT NOT NULL,
			resource TEXT NOT NULL,
			details TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		// Score Adjustments
		`CREATE TABLE IF NOT EXISTS score_adjustments (
			id TEXT PRIMARY KEY,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			delta INTEGER NOT NULL,
			reason TEXT,
			created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TEXT NOT NULL
		);`,
		// Contest Config
		`CREATE TABLE IF NOT EXISTS contest_config (
			id TEXT PRIMARY KEY,
			contest_id TEXT REFERENCES contests(id) ON DELETE SET NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			freeze_time TEXT,
			title TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 0,
			is_paused INTEGER NOT NULL DEFAULT 0,
			scoreboard_visibility TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Contests
		`CREATE TABLE IF NOT EXISTS contests (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			freeze_time TEXT,
			scoreboard_visibility TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Contest Rounds
		`CREATE TABLE IF NOT EXISTS contest_rounds (
			id TEXT PRIMARY KEY,
			contest_id TEXT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			display_order INTEGER NOT NULL,
			visible_from TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		// Round Challenges (Junction)
		`CREATE TABLE IF NOT EXISTS round_challenges (
			id TEXT PRIMARY KEY,
			round_id TEXT NOT NULL REFERENCES contest_rounds(id) ON DELETE CASCADE,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			created_at TEXT NOT NULL,
			UNIQUE(round_id, challenge_id)
		);`,
		// Team Contest Registrations
		`CREATE TABLE IF NOT EXISTS team_contest_registrations (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			contest_id TEXT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
			registered_at TEXT NOT NULL,
			UNIQUE(team_id, contest_id)
		);`,
		// Contest Challenge Solves (per-contest solve counts for dynamic scoring isolation)
		`CREATE TABLE IF NOT EXISTS contest_challenge_solves (
			contest_id TEXT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
			challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
			solve_count INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (contest_id, challenge_id)
		);`,
		// User IP History
		`CREATE TABLE IF NOT EXISTS user_ip_history (
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			ip TEXT NOT NULL,
			action TEXT,
			timestamp TEXT NOT NULL
		);`,
	}

	for _, stmt := range schemas {
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("Failed to execute schema statement: %v\nQuery: %s", err, stmt)
		}
	}
	log.Println("Database schema bootstrapped successfully with consistent relations")
}
