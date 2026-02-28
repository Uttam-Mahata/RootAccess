package repositories

import (
	"context"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TeamContestRegistrationRepository struct {
	collection *mongo.Collection
}

func NewTeamContestRegistrationRepository() *TeamContestRegistrationRepository {
	return &TeamContestRegistrationRepository{
		collection: database.DB.Collection("team_contest_registrations"),
	}
}

// CreateIndexes creates the unique compound index on (contest_id, team_id).
func (r *TeamContestRegistrationRepository) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "contest_id", Value: 1}, {Key: "team_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// RegisterTeam registers a team for a contest
func (r *TeamContestRegistrationRepository) RegisterTeam(teamID, contestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if already registered
	filter := bson.M{
		"team_id":    teamID,
		"contest_id": contestID,
	}
	var existing models.TeamContestRegistration
	err := r.collection.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		// Already registered
		return nil
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// Create registration
	registration := models.TeamContestRegistration{
		TeamID:       teamID,
		ContestID:    contestID,
		RegisteredAt: time.Now(),
	}
	_, err = r.collection.InsertOne(ctx, registration)
	return err
}

// UnregisterTeam unregisters a team from a contest
func (r *TeamContestRegistrationRepository) UnregisterTeam(teamID, contestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"team_id":    teamID,
		"contest_id": contestID,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// IsTeamRegistered checks if a team is registered for a contest
func (r *TeamContestRegistrationRepository) IsTeamRegistered(teamID, contestID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"team_id":    teamID,
		"contest_id": contestID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	return count > 0, err
}

// GetTeamContests returns all contest IDs a team is registered for
func (r *TeamContestRegistrationRepository) GetTeamContests(teamID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"team_id": teamID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var registrations []models.TeamContestRegistration
	if err := cursor.All(ctx, &registrations); err != nil {
		return nil, err
	}

	contestIDs := make([]primitive.ObjectID, len(registrations))
	for i, reg := range registrations {
		contestIDs[i] = reg.ContestID
	}
	return contestIDs, nil
}

// CountContestTeams returns the number of teams registered for a contest
func (r *TeamContestRegistrationRepository) CountContestTeams(contestID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"contest_id": contestID})
}

// GetContestTeams returns all team IDs registered for a contest
func (r *TeamContestRegistrationRepository) GetContestTeams(contestID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"contest_id": contestID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var registrations []models.TeamContestRegistration
	if err := cursor.All(ctx, &registrations); err != nil {
		return nil, err
	}

	teamIDs := make([]primitive.ObjectID, len(registrations))
	for i, reg := range registrations {
		teamIDs[i] = reg.TeamID
	}
	return teamIDs, nil
}
