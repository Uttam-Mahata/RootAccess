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

type ContestRoundRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewContestRoundRepository(db *mongo.Database) *ContestRoundRepository {
	return &ContestRoundRepository{
		db:         db,
		collection: db.Collection("contest_rounds"),
	}
}

func (r *ContestRoundRepository) Create(round *models.ContestRound) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	round.CreatedAt = time.Now()
	round.UpdatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, round)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		round.ID = oid
	}
	return nil
}

func (r *ContestRoundRepository) Update(round *models.ContestRound) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	round.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": round.ID}, round)
	return err
}

func (r *ContestRoundRepository) FindByID(id string) (*models.ContestRound, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var round models.ContestRound
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&round)
	if err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *ContestRoundRepository) ListByContestID(contestID string) ([]models.ContestRound, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "start_time", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"contest_id": oid}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rounds []models.ContestRound
	if err = cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}
	return rounds, nil
}

// GetActiveRounds returns rounds that are visible and active at the given time
func (r *ContestRoundRepository) GetActiveRounds(contestID string, now time.Time) ([]models.ContestRound, error) {
	rounds, err := r.ListByContestID(contestID)
	if err != nil {
		return nil, err
	}

	var active []models.ContestRound
	for _, round := range rounds {
		if round.IsRoundVisibleAt(now) {
			active = append(active, round)
		}
	}
	return active, nil
}

func (r *ContestRoundRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *ContestRoundRepository) DeleteByContestID(contestID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"contest_id": oid})
	return err
}
