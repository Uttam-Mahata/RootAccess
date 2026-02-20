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

// GetChallengesByIDs returns challenges with the given IDs that are published
func (r *ChallengeRepository) GetChallengesByIDs(ids []primitive.ObjectID, publishedOnly bool) ([]models.Challenge, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$in": ids}}
	if publishedOnly {
		filter["is_published"] = true
	}

	cursor, err := r.collection.Find(ctx, filter)
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

// UpdateOfficialWriteup updates the official writeup fields on a challenge
func (r *ChallengeRepository) UpdateOfficialWriteup(id string, content, format string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{
			"official_writeup":         content,
			"official_writeup_format":  format,
		},
	})
	return err
}

// SetContestID sets the contest_id on a challenge
func (r *ChallengeRepository) SetContestID(id string, contestID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{"contest_id": contestID},
	})
	return err
}

// PublishOfficialWriteup sets OfficialWriteupPublished to true
func (r *ChallengeRepository) PublishOfficialWriteup(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{"official_writeup_published": true},
	})
	return err
}
