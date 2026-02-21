package mongodb

import (
	"context"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubmissionRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewSubmissionRepository(db *mongo.Database) *SubmissionRepository {
	return &SubmissionRepository{
		db:         db,
		collection: db.Collection("submissions"),
	}
}

func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	submission.Timestamp = time.Now()
	result, err := r.collection.InsertOne(ctx, submission)
	if err != nil {
		return err
	}
	// Populate the generated ID back to the submission struct
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		submission.ID = oid
	}
	return nil
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

// CountSubmissions returns the total number of submissions
func (r *SubmissionRepository) CountSubmissions() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountCorrectSubmissions returns the total number of correct submissions
func (r *SubmissionRepository) CountCorrectSubmissions() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"is_correct": true})
}

// GetAllSubmissions returns all submissions
func (r *SubmissionRepository) GetAllSubmissions() ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
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

// GetRecentSubmissions returns the most recent submissions, limited by count
func (r *SubmissionRepository) GetRecentSubmissions(limit int64) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
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

// GetCorrectSubmissionsSince returns correct submissions since a given time
func (r *SubmissionRepository) GetCorrectSubmissionsSince(since time.Time) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_correct": true,
		"timestamp":  bson.M{"$gte": since},
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// GetSubmissionsSince returns all submissions since a given time
func (r *SubmissionRepository) GetSubmissionsSince(since time.Time) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"timestamp": bson.M{"$gte": since},
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// GetCorrectSubmissionsByChallenge returns all correct submissions for a specific challenge
func (r *SubmissionRepository) GetCorrectSubmissionsByChallenge(challengeID primitive.ObjectID) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"challenge_id": challengeID,
		"is_correct":   true,
	}, opts)
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

// GetCorrectSubmissionsBefore returns correct submissions before a given time (for scoreboard freeze)
func (r *SubmissionRepository) GetCorrectSubmissionsBefore(before time.Time) ([]models.Submission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_correct": true,
		"timestamp":  bson.M{"$lte": before},
	}

	cursor, err := r.collection.Find(ctx, filter)
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
