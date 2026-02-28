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

type HintRepository struct {
	revealCollection *mongo.Collection
}

func NewHintRepository() *HintRepository {
	return &HintRepository{
		revealCollection: database.DB.Collection("hint_reveals"),
	}
}

// FindReveal checks if a user/team has already revealed a hint
func (r *HintRepository) FindReveal(hintID, userID primitive.ObjectID) (*models.HintReveal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reveal models.HintReveal
	err := r.revealCollection.FindOne(ctx, bson.M{
		"hint_id": hintID,
		"user_id": userID,
	}).Decode(&reveal)
	if err != nil {
		return nil, err
	}
	return &reveal, nil
}

// FindRevealByTeam checks if a team member has already revealed a hint
func (r *HintRepository) FindRevealByTeam(hintID, teamID primitive.ObjectID) (*models.HintReveal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reveal models.HintReveal
	err := r.revealCollection.FindOne(ctx, bson.M{
		"hint_id": hintID,
		"team_id": teamID,
	}).Decode(&reveal)
	if err != nil {
		return nil, err
	}
	return &reveal, nil
}

// CreateReveal records that a hint was revealed
func (r *HintRepository) CreateReveal(reveal *models.HintReveal) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.revealCollection.InsertOne(ctx, reveal)
	return err
}

// GetRevealsByUserAndChallenge gets all reveals for a user on a challenge
func (r *HintRepository) GetRevealsByUserAndChallenge(userID, challengeID primitive.ObjectID) ([]models.HintReveal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.revealCollection.Find(ctx, bson.M{
		"user_id":      userID,
		"challenge_id": challengeID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reveals []models.HintReveal
	if err = cursor.All(ctx, &reveals); err != nil {
		return nil, err
	}
	return reveals, nil
}

// GetRevealsByTeamAndChallenge gets all reveals for a team on a challenge
func (r *HintRepository) GetRevealsByTeamAndChallenge(teamID, challengeID primitive.ObjectID) ([]models.HintReveal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.revealCollection.Find(ctx, bson.M{
		"team_id":      teamID,
		"challenge_id": challengeID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reveals []models.HintReveal
	if err = cursor.All(ctx, &reveals); err != nil {
		return nil, err
	}
	return reveals, nil
}
