package repositories

import (
	"context"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SubmissionRepository struct {
	collection *mongo.Collection
}

func NewSubmissionRepository() *SubmissionRepository {
	return &SubmissionRepository{
		collection: database.DB.Collection("submissions"),
	}
}

func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	submission.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, submission)
	return err
}

func (r *SubmissionRepository) FindByChallengeAndUser(challengeID, userID primitive.ObjectID) (*models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var submission models.Submission
	err := r.collection.FindOne(ctx, bson.M{
		"challenge_id": challengeID,
		"user_id":      userID,
		"is_correct":   true,
	}).Decode(&submission)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) FindByChallengeAndTeam(challengeID, teamID primitive.ObjectID) (*models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var submission models.Submission
	err := r.collection.FindOne(ctx, bson.M{
		"challenge_id": challengeID,
		"team_id":      teamID,
		"is_correct":   true,
	}).Decode(&submission)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) GetTeamSubmissions(teamID primitive.ObjectID) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"team_id":    teamID,
		"is_correct": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (r *SubmissionRepository) GetAllCorrectSubmissions() ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"is_correct": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}
	return submissions, nil
}

// GetUserCorrectSubmissions returns all correct submissions by a specific user
func (r *SubmissionRepository) GetUserCorrectSubmissions(userID primitive.ObjectID) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":    userID,
		"is_correct": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}
	return submissions, nil
}

// GetUserSubmissionCount returns the total number of submissions by a user
func (r *SubmissionRepository) GetUserSubmissionCount(userID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
}

// GetUserCorrectSubmissionCount returns the number of correct submissions by a user
func (r *SubmissionRepository) GetUserCorrectSubmissionCount(userID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"user_id": userID, "is_correct": true})
}
