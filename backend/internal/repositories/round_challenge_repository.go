package repositories

import (
	"context"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoundChallengeRepository struct {
	collection *mongo.Collection
}

func NewRoundChallengeRepository() *RoundChallengeRepository {
	return &RoundChallengeRepository{
		collection: database.DB.Collection("round_challenges"),
	}
}

func (r *RoundChallengeRepository) Attach(roundID, challengeID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Avoid duplicate
	filter := bson.M{"round_id": roundID, "challenge_id": challengeID}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already attached
	}

	rc := &models.RoundChallenge{
		RoundID:     roundID,
		ChallengeID: challengeID,
		CreatedAt:   time.Now(),
	}
	_, err = r.collection.InsertOne(ctx, rc)
	return err
}

func (r *RoundChallengeRepository) Detach(roundID, challengeID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"round_id": roundID, "challenge_id": challengeID})
	return err
}

func (r *RoundChallengeRepository) GetChallengeIDsForRounds(roundIDs []primitive.ObjectID) ([]primitive.ObjectID, error) {
	if len(roundIDs) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"round_id": bson.M{"$in": roundIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	seen := make(map[string]bool)
	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var rc models.RoundChallenge
		if err := cursor.Decode(&rc); err != nil {
			return nil, err
		}
		hex := rc.ChallengeID.Hex()
		if !seen[hex] {
			seen[hex] = true
			ids = append(ids, rc.ChallengeID)
		}
	}
	return ids, cursor.Err()
}

func (r *RoundChallengeRepository) GetRoundIDsForChallenge(challengeID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"challenge_id": challengeID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var rc models.RoundChallenge
		if err := cursor.Decode(&rc); err != nil {
			return nil, err
		}
		ids = append(ids, rc.RoundID)
	}
	return ids, cursor.Err()
}

func (r *RoundChallengeRepository) GetChallengesByRound(roundID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"round_id": roundID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var rc models.RoundChallenge
		if err := cursor.Decode(&rc); err != nil {
			return nil, err
		}
		ids = append(ids, rc.ChallengeID)
	}
	return ids, cursor.Err()
}
