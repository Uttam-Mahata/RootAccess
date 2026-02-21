package mongodb

import (
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
)

// Compile-time interface compliance checks.
var (
	_ interfaces.UserRepository                  = (*UserRepository)(nil)
	_ interfaces.ChallengeRepository             = (*ChallengeRepository)(nil)
	_ interfaces.SubmissionRepository             = (*SubmissionRepository)(nil)
	_ interfaces.TeamRepository                  = (*TeamRepository)(nil)
	_ interfaces.TeamInvitationRepository        = (*TeamInvitationRepository)(nil)
	_ interfaces.NotificationRepository          = (*NotificationRepository)(nil)
	_ interfaces.HintRepository                  = (*HintRepository)(nil)
	_ interfaces.ContestRepository               = (*ContestRepository)(nil)
	_ interfaces.ContestEntityRepository         = (*ContestEntityRepository)(nil)
	_ interfaces.ContestRoundRepository          = (*ContestRoundRepository)(nil)
	_ interfaces.RoundChallengeRepository        = (*RoundChallengeRepository)(nil)
	_ interfaces.WriteupRepository               = (*WriteupRepository)(nil)
	_ interfaces.AuditLogRepository              = (*AuditLogRepository)(nil)
	_ interfaces.AchievementRepository           = (*AchievementRepository)(nil)
	_ interfaces.ScoreAdjustmentRepository       = (*ScoreAdjustmentRepository)(nil)
	_ interfaces.TeamContestRegistrationRepository = (*TeamContestRegistrationRepository)(nil)
)
