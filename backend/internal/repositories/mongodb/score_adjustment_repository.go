package mongodb

import (
	"context"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScoreAdjustmentRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewScoreAdjustmentRepository(db *mongo.Database) *ScoreAdjustmentRepository {
	return &ScoreAdjustmentRepository{
		db:         db,
		collection: db.Collection("score_adjustments"),
	}
}

func (r *ScoreAdjustmentRepository) Create(adjustment *models.ScoreAdjustment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adjustment.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, adjustment)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		adjustment.ID = oid
	}
	return nil
}

// GetAdjustmentsForUsers returns a map of userID(hex) -> total delta.
func (r *ScoreAdjustmentRepository) GetAdjustmentsForUsers(userIDs []primitive.ObjectID) (map[string]int, error) {
	return r.getAdjustmentsByTargets(models.ScoreAdjustmentTargetUser, userIDs)
}

// GetAdjustmentsForTeams returns a map of teamID(hex) -> total delta.
func (r *ScoreAdjustmentRepository) GetAdjustmentsForTeams(teamIDs []primitive.ObjectID) (map[string]int, error) {
	return r.getAdjustmentsByTargets(models.ScoreAdjustmentTargetTeam, teamIDs)
}

func (r *ScoreAdjustmentRepository) getAdjustmentsByTargets(targetType string, ids []primitive.ObjectID) (map[string]int, error) {
	result := make(map[string]int)
	if len(ids) == 0 {
		return result, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"target_type": targetType,
		"target_id": bson.M{
			"$in": ids,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var adj models.ScoreAdjustment
		if err := cursor.Decode(&adj); err != nil {
			return nil, err
		}
		key := adj.TargetID.Hex()
		result[key] += adj.Delta
	}

	return result, cursor.Err()
}

