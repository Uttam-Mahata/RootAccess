package interfaces

import (
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines methods for user data access.
type UserRepository interface {
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	FindByID(userID string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByVerificationToken(token string) (*models.User, error)
	FindByResetToken(token string) (*models.User, error)
	FindByProviderID(provider, providerID string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateFields(userID primitive.ObjectID, fields map[string]interface{}) error
	CountUsers() (int64, error)
	RecordUserIP(userID primitive.ObjectID, ip string, action string) error
	GetUsersWithDetails() ([]models.User, error)
	CountUsersByStatus(status string) (int64, error)
	CountVerifiedUsers() (int64, error)
	CountAdmins() (int64, error)
	GetRecentUsers(since time.Time) ([]models.User, error)
}

// ChallengeRepository defines methods for challenge data access.
type ChallengeRepository interface {
	CreateChallenge(challenge *models.Challenge) error
	GetAllChallenges() ([]models.Challenge, error)
	GetAllChallengesForList() ([]models.Challenge, error)
	GetChallengeByID(id string) (*models.Challenge, error)
	UpdateChallenge(id string, challenge *models.Challenge) error
	DeleteChallenge(id string) error
	IncrementSolveCount(id string) error
	GetFlagHash(id string) (string, error)
	CountChallenges() (int64, error)
	GetChallengesByIDs(ids []primitive.ObjectID, publishedOnly bool) ([]models.Challenge, error)
	UpdateOfficialWriteup(id string, content string, format string) error
	SetContestID(id string, contestID primitive.ObjectID) error
	PublishOfficialWriteup(id string) error
}

// SubmissionRepository defines methods for submission data access.
type SubmissionRepository interface {
	CreateSubmission(submission *models.Submission) error
	FindByChallengeAndUser(challengeID, userID primitive.ObjectID) (*models.Submission, error)
	FindByChallengeAndTeam(challengeID, teamID primitive.ObjectID) (*models.Submission, error)
	GetTeamSubmissions(teamID primitive.ObjectID) ([]models.Submission, error)
	GetAllCorrectSubmissions() ([]models.Submission, error)
	GetUserCorrectSubmissions(userID primitive.ObjectID) ([]models.Submission, error)
	GetUserSubmissionCount(userID primitive.ObjectID) (int64, error)
	GetUserCorrectSubmissionCount(userID primitive.ObjectID) (int64, error)
	CountSubmissions() (int64, error)
	CountCorrectSubmissions() (int64, error)
	GetAllSubmissions() ([]models.Submission, error)
	GetRecentSubmissions(limit int64) ([]models.Submission, error)
	GetCorrectSubmissionsSince(since time.Time) ([]models.Submission, error)
	GetSubmissionsSince(since time.Time) ([]models.Submission, error)
	GetCorrectSubmissionsByChallenge(challengeID primitive.ObjectID) ([]models.Submission, error)
	GetCorrectSubmissionsBefore(before time.Time) ([]models.Submission, error)
}

// TeamRepository defines methods for team data access.
type TeamRepository interface {
	CreateTeam(team *models.Team) error
	FindTeamByID(teamID string) (*models.Team, error)
	FindTeamByLeaderID(leaderID string) (*models.Team, error)
	FindTeamByMemberID(userID string) (*models.Team, error)
	FindTeamByInviteCode(code string) (*models.Team, error)
	FindTeamByName(name string) (*models.Team, error)
	UpdateTeam(team *models.Team) error
	DeleteTeam(teamID string) error
	AddMemberToTeam(teamID, userID string) error
	RemoveMemberFromTeam(teamID, userID string) error
	UpdateTeamScore(teamID string, points int) error
	GetAllTeamsWithScores() ([]models.Team, error)
	CountTeams() (int64, error)
	GetTeamMemberCount(teamID string) (int, error)
	GetAllTeams() ([]models.Team, error)
	UpdateTeamFields(teamID primitive.ObjectID, fields map[string]interface{}) error
	AdminDeleteTeam(teamID primitive.ObjectID) error
	AdminUpdateTeamLeader(teamID, newLeaderID primitive.ObjectID) error
	GetRecentTeams(since time.Time) ([]models.Team, error)
}

// TeamInvitationRepository defines methods for team invitation data access.
type TeamInvitationRepository interface {
	CreateInvitation(invitation *models.TeamInvitation) error
	FindInvitationByID(invitationID string) (*models.TeamInvitation, error)
	FindInvitationByToken(token string) (*models.TeamInvitation, error)
	FindPendingInvitationsForUser(userID, email string) ([]models.TeamInvitation, error)
	FindInvitationsByTeam(teamID string) ([]models.TeamInvitation, error)
	FindPendingInvitationsByTeam(teamID string) ([]models.TeamInvitation, error)
	UpdateInvitationStatus(invitationID, status string) error
	DeleteExpiredInvitations() error
	DeleteInvitationsByTeam(teamID string) error
	HasPendingInvitation(teamID, userID, email string) (bool, error)
}

// NotificationRepository defines methods for notification data access.
type NotificationRepository interface {
	CreateNotification(notification *models.Notification) error
	GetAllNotifications() ([]models.Notification, error)
	GetActiveNotifications() ([]models.Notification, error)
	GetNotificationByID(id string) (*models.Notification, error)
	UpdateNotification(id string, notification *models.Notification) error
	DeleteNotification(id string) error
	ToggleNotificationActive(id string) error
}

// HintRepository defines methods for hint reveal data access.
type HintRepository interface {
	FindReveal(hintID, userID primitive.ObjectID) (*models.HintReveal, error)
	FindRevealByTeam(hintID, teamID primitive.ObjectID) (*models.HintReveal, error)
	CreateReveal(reveal *models.HintReveal) error
	GetRevealsByUserAndChallenge(userID, challengeID primitive.ObjectID) ([]models.HintReveal, error)
	GetRevealsByTeamAndChallenge(teamID, challengeID primitive.ObjectID) ([]models.HintReveal, error)
}

// ContestRepository defines methods for contest config data access.
type ContestRepository interface {
	GetActiveContest() (*models.ContestConfig, error)
	UpsertContest(config *models.ContestConfig) error
}

// ContestEntityRepository defines methods for contest entity data access.
type ContestEntityRepository interface {
	Create(contest *models.Contest) error
	Update(contest *models.Contest) error
	FindByID(id string) (*models.Contest, error)
	ListAll() ([]models.Contest, error)
	GetScoreboardContests() ([]models.Contest, error)
	Delete(id string) error
}

// ContestRoundRepository defines methods for contest round data access.
type ContestRoundRepository interface {
	Create(round *models.ContestRound) error
	Update(round *models.ContestRound) error
	FindByID(id string) (*models.ContestRound, error)
	ListByContestID(contestID string) ([]models.ContestRound, error)
	GetActiveRounds(contestID string, now time.Time) ([]models.ContestRound, error)
	Delete(id string) error
	DeleteByContestID(contestID string) error
}

// RoundChallengeRepository defines methods for round-challenge association data access.
type RoundChallengeRepository interface {
	Attach(roundID, challengeID primitive.ObjectID) error
	Detach(roundID, challengeID primitive.ObjectID) error
	GetChallengeIDsForRounds(roundIDs []primitive.ObjectID) ([]primitive.ObjectID, error)
	GetRoundIDsForChallenge(challengeID primitive.ObjectID) ([]primitive.ObjectID, error)
	GetChallengesByRound(roundID primitive.ObjectID) ([]primitive.ObjectID, error)
}

// WriteupRepository defines methods for writeup data access.
type WriteupRepository interface {
	CreateWriteup(writeup *models.Writeup) error
	GetWriteupByID(id string) (*models.Writeup, error)
	GetWriteupsByChallenge(challengeID primitive.ObjectID, onlyApproved bool) ([]models.Writeup, error)
	GetWriteupsByUser(userID primitive.ObjectID) ([]models.Writeup, error)
	GetAllWriteups() ([]models.Writeup, error)
	UpdateWriteupStatus(id string, status string) error
	DeleteWriteup(id string) error
	FindByUserAndChallenge(userID, challengeID primitive.ObjectID) (*models.Writeup, error)
	UpdateWriteupContent(id string, content string, contentFormat string) error
	GetWriteupsByTeam(teamID string) ([]models.Writeup, error)
	ToggleUpvote(id string, userID primitive.ObjectID) (bool, error)
}

// AuditLogRepository defines methods for audit log data access.
type AuditLogRepository interface {
	CreateLog(log *models.AuditLog) error
	GetLogs(limit int, page int) ([]models.AuditLog, error)
	GetLogCount() (int64, error)
}

// AchievementRepository defines methods for achievement data access.
type AchievementRepository interface {
	Create(achievement *models.Achievement) error
	GetByUserID(userID primitive.ObjectID) ([]models.Achievement, error)
	GetByTeamID(teamID primitive.ObjectID) ([]models.Achievement, error)
	GetByType(achievementType string) ([]models.Achievement, error)
	Exists(userID primitive.ObjectID, achievementType string) (bool, error)
	ExistsForCategory(userID primitive.ObjectID, achievementType string, category string) (bool, error)
}

// ScoreAdjustmentRepository defines methods for score adjustment data access.
type ScoreAdjustmentRepository interface {
	Create(adjustment *models.ScoreAdjustment) error
	GetAdjustmentsForUsers(userIDs []primitive.ObjectID) (map[string]int, error)
	GetAdjustmentsForTeams(teamIDs []primitive.ObjectID) (map[string]int, error)
}

// TeamContestRegistrationRepository defines methods for team contest registration data access.
type TeamContestRegistrationRepository interface {
	CreateIndexes() error
	RegisterTeam(teamID, contestID primitive.ObjectID) error
	UnregisterTeam(teamID, contestID primitive.ObjectID) error
	IsTeamRegistered(teamID, contestID primitive.ObjectID) (bool, error)
	GetTeamContests(teamID primitive.ObjectID) ([]primitive.ObjectID, error)
	CountContestTeams(contestID primitive.ObjectID) (int64, error)
	GetContestTeams(contestID primitive.ObjectID) ([]primitive.ObjectID, error)
}
