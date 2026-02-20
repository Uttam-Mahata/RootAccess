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
