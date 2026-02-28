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

type TeamInvitationRepository struct {
	collection *mongo.Collection
}

func NewTeamInvitationRepository() *TeamInvitationRepository {
	return &TeamInvitationRepository{
		collection: database.DB.Collection("team_invitations"),
	}
}

func (r *TeamInvitationRepository) CreateInvitation(invitation *models.TeamInvitation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	invitation.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, invitation)
	if err != nil {
		return err
	}
	invitation.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TeamInvitationRepository) FindInvitationByID(invitationID string) (*models.TeamInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(invitationID)
	if err != nil {
		return nil, err
	}

	var invitation models.TeamInvitation
	err = r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&invitation)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *TeamInvitationRepository) FindInvitationByToken(token string) (*models.TeamInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var invitation models.TeamInvitation
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&invitation)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *TeamInvitationRepository) FindPendingInvitationsForUser(userID, email string) ([]models.TeamInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var filter bson.M
	if userID != "" {
		userObjID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return nil, err
		}
		if email != "" {
			filter = bson.M{
				"status": models.InvitationStatusPending,
				"$or": []bson.M{
					{"invitee_user_id": userObjID},
					{"invitee_email": email},
				},
			}
		} else {
			filter = bson.M{
				"status":          models.InvitationStatusPending,
				"invitee_user_id": userObjID,
			}
		}
	} else if email != "" {
		filter = bson.M{
			"status":        models.InvitationStatusPending,
			"invitee_email": email,
		}
	} else {
		return []models.TeamInvitation{}, nil
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []models.TeamInvitation
	if err = cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *TeamInvitationRepository) FindInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"team_id": teamObjID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []models.TeamInvitation
	if err = cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *TeamInvitationRepository) FindPendingInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"team_id": teamObjID,
		"status":  models.InvitationStatusPending,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []models.TeamInvitation
	if err = cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *TeamInvitationRepository) UpdateInvitationStatus(invitationID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(invitationID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status}}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *TeamInvitationRepository) DeleteExpiredInvitations() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"status":     models.InvitationStatusPending,
		"expires_at": bson.M{"$lt": time.Now()},
	}
	update := bson.M{"$set": bson.M{"status": models.InvitationStatusExpired}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *TeamInvitationRepository) DeleteInvitationsByTeam(teamID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"team_id": teamObjID})
	return err
}

func (r *TeamInvitationRepository) HasPendingInvitation(teamID, userID, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return false, err
	}

	var filter bson.M
	if userID != "" {
		userObjID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return false, err
		}
		filter = bson.M{
			"team_id":         teamObjID,
			"status":          models.InvitationStatusPending,
			"invitee_user_id": userObjID,
		}
	} else if email != "" {
		filter = bson.M{
			"team_id":       teamObjID,
			"status":        models.InvitationStatusPending,
			"invitee_email": email,
		}
	} else {
		return false, nil
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
