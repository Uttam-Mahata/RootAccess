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

type WriteupRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewWriteupRepository(db *mongo.Database) *WriteupRepository {
	return &WriteupRepository{
		db:         db,
		collection: db.Collection("writeups"),
	}
}

func (r *WriteupRepository) CreateWriteup(writeup *models.Writeup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	writeup.CreatedAt = time.Now()
	writeup.UpdatedAt = time.Now()
	// Only set status to pending if not already set (allows auto-approval)
	if writeup.Status == "" {
		writeup.Status = models.WriteupStatusPending
	}

	_, err := r.collection.InsertOne(ctx, writeup)
	return err
}

func (r *WriteupRepository) GetWriteupByID(id string) (*models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var writeup models.Writeup
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&writeup)
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *WriteupRepository) GetWriteupsByChallenge(challengeID primitive.ObjectID, onlyApproved bool) ([]models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"challenge_id": challengeID}
	if onlyApproved {
		filter["status"] = models.WriteupStatusApproved
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var writeups []models.Writeup
	if err = cursor.All(ctx, &writeups); err != nil {
		return nil, err
	}
	return writeups, nil
}

func (r *WriteupRepository) GetWriteupsByUser(userID primitive.ObjectID) ([]models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var writeups []models.Writeup
	if err = cursor.All(ctx, &writeups); err != nil {
		return nil, err
	}
	return writeups, nil
}

func (r *WriteupRepository) GetAllWriteups() ([]models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var writeups []models.Writeup
	if err = cursor.All(ctx, &writeups); err != nil {
		return nil, err
	}
	return writeups, nil
}

func (r *WriteupRepository) UpdateWriteupStatus(id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *WriteupRepository) DeleteWriteup(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// FindByUserAndChallenge checks if a user already submitted a writeup for a challenge
func (r *WriteupRepository) FindByUserAndChallenge(userID, challengeID primitive.ObjectID) (*models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var writeup models.Writeup
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":      userID,
		"challenge_id": challengeID,
	}).Decode(&writeup)
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

// UpdateWriteupContent updates the content and format of a writeup
func (r *WriteupRepository) UpdateWriteupContent(id string, content string, contentFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"content":        content,
			"content_format": contentFormat,
			"updated_at":     time.Now(),
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// GetWriteupsByTeam returns all writeups for a specific team
func (r *WriteupRepository) GetWriteupsByTeam(teamID string) ([]models.Writeup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tid, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	// Get all users in the team first
	teamCollection := r.db.Collection("teams")
	var team models.Team
	err = teamCollection.FindOne(ctx, bson.M{"_id": tid}).Decode(&team)
	if err != nil {
		return nil, err
	}

	// Get writeups for all team members
	filter := bson.M{"user_id": bson.M{"$in": team.MemberIDs}}
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var writeups []models.Writeup
	if err = cursor.All(ctx, &writeups); err != nil {
		return nil, err
	}
	return writeups, nil
}

// ToggleUpvote adds or removes a user's upvote on a writeup
func (r *WriteupRepository) ToggleUpvote(id string, userID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	// Try to add the upvote atomically using $addToSet (prevents duplicates)
	addResult, err := r.collection.UpdateOne(ctx, bson.M{
		"_id":        oid,
		"upvoted_by": bson.M{"$ne": userID},
	}, bson.M{
		"$addToSet": bson.M{"upvoted_by": userID},
		"$inc":      bson.M{"upvotes": 1},
	})
	if err != nil {
		return false, err
	}

	// If the add succeeded, user was not previously in the list
	if addResult.ModifiedCount > 0 {
		return true, nil
	}

	// User already upvoted, so remove the upvote
	_, err = r.collection.UpdateOne(ctx, bson.M{
		"_id":        oid,
		"upvoted_by": userID,
	}, bson.M{
		"$pull": bson.M{"upvoted_by": userID},
		"$inc":  bson.M{"upvotes": -1},
	})
	return false, err
}
