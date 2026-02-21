package repositories

import (
	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/mongodb"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/turso"
)

// isTurso returns true when the Turso SQL backend is active.
func isTurso() bool {
	return database.TursoDB != nil
}

// NewUserRepository creates a UserRepository based on the configured DB type.
func NewUserRepository() interfaces.UserRepository {
	if isTurso() {
		return turso.NewUserRepository(database.TursoDB)
	}
	return mongodb.NewUserRepository(database.DB)
}

// NewChallengeRepository creates a ChallengeRepository based on the configured DB type.
func NewChallengeRepository() interfaces.ChallengeRepository {
	if isTurso() {
		return turso.NewChallengeRepository(database.TursoDB)
	}
	return mongodb.NewChallengeRepository(database.DB)
}

// NewSubmissionRepository creates a SubmissionRepository based on the configured DB type.
func NewSubmissionRepository() interfaces.SubmissionRepository {
	if isTurso() {
		return turso.NewSubmissionRepository(database.TursoDB)
	}
	return mongodb.NewSubmissionRepository(database.DB)
}

// NewTeamRepository creates a TeamRepository based on the configured DB type.
func NewTeamRepository() interfaces.TeamRepository {
	if isTurso() {
		return turso.NewTeamRepository(database.TursoDB)
	}
	return mongodb.NewTeamRepository(database.DB)
}

// NewTeamInvitationRepository creates a TeamInvitationRepository based on the configured DB type.
func NewTeamInvitationRepository() interfaces.TeamInvitationRepository {
	if isTurso() {
		return turso.NewTeamInvitationRepository(database.TursoDB)
	}
	return mongodb.NewTeamInvitationRepository(database.DB)
}

// NewNotificationRepository creates a NotificationRepository based on the configured DB type.
func NewNotificationRepository() interfaces.NotificationRepository {
	if isTurso() {
		return turso.NewNotificationRepository(database.TursoDB)
	}
	return mongodb.NewNotificationRepository(database.DB)
}

// NewHintRepository creates a HintRepository based on the configured DB type.
func NewHintRepository() interfaces.HintRepository {
	if isTurso() {
		return turso.NewHintRepository(database.TursoDB)
	}
	return mongodb.NewHintRepository(database.DB)
}

// NewContestRepository creates a ContestRepository based on the configured DB type.
func NewContestRepository() interfaces.ContestRepository {
	if isTurso() {
		return turso.NewContestRepository(database.TursoDB)
	}
	return mongodb.NewContestRepository(database.DB)
}

// NewContestEntityRepository creates a ContestEntityRepository based on the configured DB type.
func NewContestEntityRepository() interfaces.ContestEntityRepository {
	if isTurso() {
		return turso.NewContestEntityRepository(database.TursoDB)
	}
	return mongodb.NewContestEntityRepository(database.DB)
}

// NewContestRoundRepository creates a ContestRoundRepository based on the configured DB type.
func NewContestRoundRepository() interfaces.ContestRoundRepository {
	if isTurso() {
		return turso.NewContestRoundRepository(database.TursoDB)
	}
	return mongodb.NewContestRoundRepository(database.DB)
}

// NewRoundChallengeRepository creates a RoundChallengeRepository based on the configured DB type.
func NewRoundChallengeRepository() interfaces.RoundChallengeRepository {
	if isTurso() {
		return turso.NewRoundChallengeRepository(database.TursoDB)
	}
	return mongodb.NewRoundChallengeRepository(database.DB)
}

// NewWriteupRepository creates a WriteupRepository based on the configured DB type.
func NewWriteupRepository() interfaces.WriteupRepository {
	if isTurso() {
		return turso.NewWriteupRepository(database.TursoDB)
	}
	return mongodb.NewWriteupRepository(database.DB)
}

// NewAuditLogRepository creates an AuditLogRepository based on the configured DB type.
func NewAuditLogRepository() interfaces.AuditLogRepository {
	if isTurso() {
		return turso.NewAuditLogRepository(database.TursoDB)
	}
	return mongodb.NewAuditLogRepository(database.DB)
}

// NewAchievementRepository creates an AchievementRepository based on the configured DB type.
func NewAchievementRepository() interfaces.AchievementRepository {
	if isTurso() {
		return turso.NewAchievementRepository(database.TursoDB)
	}
	return mongodb.NewAchievementRepository(database.DB)
}

// NewScoreAdjustmentRepository creates a ScoreAdjustmentRepository based on the configured DB type.
func NewScoreAdjustmentRepository() interfaces.ScoreAdjustmentRepository {
	if isTurso() {
		return turso.NewScoreAdjustmentRepository(database.TursoDB)
	}
	return mongodb.NewScoreAdjustmentRepository(database.DB)
}

// NewTeamContestRegistrationRepository creates a TeamContestRegistrationRepository based on the configured DB type.
func NewTeamContestRegistrationRepository() interfaces.TeamContestRegistrationRepository {
	if isTurso() {
		return turso.NewTeamContestRegistrationRepository(database.TursoDB)
	}
	return mongodb.NewTeamContestRegistrationRepository(database.DB)
}
