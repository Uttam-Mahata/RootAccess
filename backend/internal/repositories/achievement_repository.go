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

type AchievementRepository struct {
	collection *mongo.Collection
}

func NewAchievementRepository() *AchievementRepository {
	return &AchievementRepository{
		collection: database.DB.Collection("achievements"),
	}
}

func (r *AchievementRepository) Create(achievement *models.Achievement) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	achievement.EarnedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, achievement)
	return err
}

func (r *AchievementRepository) GetByUserID(userID primitive.ObjectID) ([]models.Achievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"earned_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}

func (r *AchievementRepository) GetByTeamID(teamID primitive.ObjectID) ([]models.Achievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"earned_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"team_id": teamID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}

func (r *AchievementRepository) GetByType(achievementType string) ([]models.Achievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"earned_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"type": achievementType}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}

// Exists checks if a specific achievement has already been granted to a user
func (r *AchievementRepository) Exists(userID primitive.ObjectID, achievementType string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"type":    achievementType,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
