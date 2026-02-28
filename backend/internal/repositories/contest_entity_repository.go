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

type ContestEntityRepository struct {
	collection *mongo.Collection
}

func NewContestEntityRepository() *ContestEntityRepository {
	return &ContestEntityRepository{
		collection: database.DB.Collection("contests"),
	}
}

func (r *ContestEntityRepository) Create(contest *models.Contest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contest.CreatedAt = time.Now()
	contest.UpdatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, contest)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		contest.ID = oid
	}
	return nil
}

func (r *ContestEntityRepository) Update(contest *models.Contest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contest.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": contest.ID}, contest)
	return err
}

func (r *ContestEntityRepository) FindByID(id string) (*models.Contest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var contest models.Contest
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&contest)
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *ContestEntityRepository) ListAll() ([]models.Contest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var contests []models.Contest
	if err = cursor.All(ctx, &contests); err != nil {
		return nil, err
	}
	return contests, nil
}

// GetScoreboardContests returns contests that should appear on the scoreboard.
// This includes contests where is_active=true AND start_time <= now (running + ended).
// Running contests are listed first, then ended contests sorted by end_time DESC.
func (r *ContestEntityRepository) GetScoreboardContests() ([]models.Contest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	filter := bson.M{
		"is_active":  true,
		"start_time": bson.M{"$lte": now},
	}

	// Sort by end_time descending so we can partition running vs ended in code
	opts := options.Find().SetSort(bson.D{{Key: "end_time", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var all []models.Contest
	if err = cursor.All(ctx, &all); err != nil {
		return nil, err
	}

	// Partition: running contests first, then ended
	var running, ended []models.Contest
	for _, c := range all {
		if c.IsRunning(now) {
			running = append(running, c)
		} else {
			ended = append(ended, c)
		}
	}

	result := make([]models.Contest, 0, len(running)+len(ended))
	result = append(result, running...)
	result = append(result, ended...)
	return result, nil
}

func (r *ContestEntityRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
