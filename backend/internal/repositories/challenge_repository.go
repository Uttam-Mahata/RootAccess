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
