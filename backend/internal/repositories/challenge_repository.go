package repositories

import (
	"context"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChallengeRepository struct {
	collection *mongo.Collection
}

func NewChallengeRepository() *ChallengeRepository {
	return &ChallengeRepository{
		collection: database.DB.Collection("challenges"),
	}
}

func (r *ChallengeRepository) CreateChallenge(challenge *models.Challenge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize solve count to 0
	challenge.SolveCount = 0

	_, err := r.collection.InsertOne(ctx, challenge)
	return err
}

func (r *ChallengeRepository) GetAllChallenges() ([]models.Challenge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []models.Challenge
	if err = cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}
	return challenges, nil
}

// GetAllChallengesForList returns challenges without description for fast list views (admin manage tab).
// Uses projection to exclude description from DB response for smaller payload and faster load.
func (r *ChallengeRepository) GetAllChallengesForList() ([]models.Challenge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.M{"description": 0})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []models.Challenge
	if err = cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}
	return challenges, nil
}

func (r *ChallengeRepository) GetChallengeByID(id string) (*models.Challenge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var challenge models.Challenge
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&challenge)
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *ChallengeRepository) UpdateChallenge(id string, challenge *models.Challenge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"title":              challenge.Title,
			"description":        challenge.Description,
			"description_format": challenge.DescriptionFormat,
			"category":           challenge.Category,
			"difficulty":         challenge.Difficulty,
			"max_points":         challenge.MaxPoints,
			"min_points":         challenge.MinPoints,
			"decay":              challenge.Decay,
			"scoring_type":       challenge.ScoringType,
			"flag_hash":          challenge.FlagHash,
			"files":              challenge.Files,
			"tags":               challenge.Tags,
			"hints":              challenge.Hints,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *ChallengeRepository) DeleteChallenge(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// IncrementSolveCount increases the solve count for a challenge by 1
func (r *ChallengeRepository) IncrementSolveCount(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{
			"solve_count": 1,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// GetFlagHash retrieves only the flag hash for verification (internal use)
func (r *ChallengeRepository) GetFlagHash(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	var result struct {
		FlagHash string `bson:"flag_hash"`
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.FlagHash, nil
}

// CountChallenges returns the total number of challenges
func (r *ChallengeRepository) CountChallenges() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

// ChallengeFilter holds the filter parameters for searching challenges
type ChallengeFilter struct {
	Category   string
	Difficulty string
	Search     string
	Tags       []string
}

// FindChallengesFiltered returns challenges matching the given filter criteria
func (r *ChallengeRepository) FindChallengesFiltered(filter ChallengeFilter) ([]models.Challenge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := bson.M{}

	if filter.Category != "" {
		query["category"] = filter.Category
	}

	if filter.Difficulty != "" {
		query["difficulty"] = filter.Difficulty
	}

	if filter.Search != "" {
		query["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": filter.Search, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$all": filter.Tags}
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []models.Challenge
	if err = cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}
	return challenges, nil
}

// GetChallengeStats returns aggregated challenge statistics
func (r *ChallengeRepository) GetChallengeStats() ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":         "$category",
			"count":       bson.M{"$sum": 1},
			"total_solves": bson.M{"$sum": "$solve_count"},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetDifficultyDistribution returns the number of challenges per difficulty level
func (r *ChallengeRepository) GetDifficultyDistribution() ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":   "$difficulty",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
