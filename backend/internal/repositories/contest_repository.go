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

type ContestRepository struct {
	collection *mongo.Collection
}

func NewContestRepository() *ContestRepository {
	return &ContestRepository{
		collection: database.DB.Collection("contest_config"),
	}
}

// GetActiveContest returns the current active contest config
func (r *ContestRepository) GetActiveContest() (*models.ContestConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config models.ContestConfig
	err := r.collection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"updated_at": -1})).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpsertContest creates or updates the contest config
func (r *ContestRepository) UpsertContest(config *models.ContestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config.UpdatedAt = time.Now()

	if config.ID.IsZero() {
		config.ID = primitive.NewObjectID()
		_, err := r.collection.InsertOne(ctx, config)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"start_time":             config.StartTime,
			"end_time":               config.EndTime,
			"freeze_time":            config.FreezeTime,
			"title":                  config.Title,
			"is_active":              config.IsActive,
			"is_paused":              config.IsPaused,
			"scoreboard_visibility":  config.ScoreboardVisibility,
			"updated_at":             config.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": config.ID}, update)
	return err
}
