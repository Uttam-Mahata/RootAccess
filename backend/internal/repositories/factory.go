package repositories

import (
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewUserRepository creates a UserRepository based on the configured DB type.
func NewUserRepository(db *mongo.Database) interfaces.UserRepository {
	return mongodb.NewUserRepository(db)
}

// NewChallengeRepository creates a ChallengeRepository based on the configured DB type.
func NewChallengeRepository(db *mongo.Database) interfaces.ChallengeRepository {
	return mongodb.NewChallengeRepository(db)
}

// NewSubmissionRepository creates a SubmissionRepository based on the configured DB type.
func NewSubmissionRepository(db *mongo.Database) interfaces.SubmissionRepository {
	return mongodb.NewSubmissionRepository(db)
}

// NewTeamRepository creates a TeamRepository based on the configured DB type.
func NewTeamRepository(db *mongo.Database) interfaces.TeamRepository {
	return mongodb.NewTeamRepository(db)
}

// NewTeamInvitationRepository creates a TeamInvitationRepository based on the configured DB type.
func NewTeamInvitationRepository(db *mongo.Database) interfaces.TeamInvitationRepository {
	return mongodb.NewTeamInvitationRepository(db)
}

// NewNotificationRepository creates a NotificationRepository based on the configured DB type.
func NewNotificationRepository(db *mongo.Database) interfaces.NotificationRepository {
	return mongodb.NewNotificationRepository(db)
}

// NewHintRepository creates a HintRepository based on the configured DB type.
func NewHintRepository(db *mongo.Database) interfaces.HintRepository {
	return mongodb.NewHintRepository(db)
}

// NewContestRepository creates a ContestRepository based on the configured DB type.
func NewContestRepository(db *mongo.Database) interfaces.ContestRepository {
	return mongodb.NewContestRepository(db)
}

// NewContestEntityRepository creates a ContestEntityRepository based on the configured DB type.
func NewContestEntityRepository(db *mongo.Database) interfaces.ContestEntityRepository {
	return mongodb.NewContestEntityRepository(db)
}

// NewContestRoundRepository creates a ContestRoundRepository based on the configured DB type.
func NewContestRoundRepository(db *mongo.Database) interfaces.ContestRoundRepository {
	return mongodb.NewContestRoundRepository(db)
}

// NewRoundChallengeRepository creates a RoundChallengeRepository based on the configured DB type.
func NewRoundChallengeRepository(db *mongo.Database) interfaces.RoundChallengeRepository {
	return mongodb.NewRoundChallengeRepository(db)
}

// NewWriteupRepository creates a WriteupRepository based on the configured DB type.
func NewWriteupRepository(db *mongo.Database) interfaces.WriteupRepository {
	return mongodb.NewWriteupRepository(db)
}

// NewAuditLogRepository creates an AuditLogRepository based on the configured DB type.
func NewAuditLogRepository(db *mongo.Database) interfaces.AuditLogRepository {
	return mongodb.NewAuditLogRepository(db)
}

// NewAchievementRepository creates an AchievementRepository based on the configured DB type.
func NewAchievementRepository(db *mongo.Database) interfaces.AchievementRepository {
	return mongodb.NewAchievementRepository(db)
}

// NewScoreAdjustmentRepository creates a ScoreAdjustmentRepository based on the configured DB type.
func NewScoreAdjustmentRepository(db *mongo.Database) interfaces.ScoreAdjustmentRepository {
	return mongodb.NewScoreAdjustmentRepository(db)
}

// NewTeamContestRegistrationRepository creates a TeamContestRegistrationRepository based on the configured DB type.
func NewTeamContestRegistrationRepository(db *mongo.Database) interfaces.TeamContestRegistrationRepository {
	return mongodb.NewTeamContestRegistrationRepository(db)
}
