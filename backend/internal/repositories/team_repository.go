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

type TeamRepository struct {
	collection *mongo.Collection
}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{
		collection: database.DB.Collection("teams"),
	}
}

func (r *TeamRepository) CreateTeam(team *models.Team) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, team)
	if err != nil {
		return err
	}
	team.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TeamRepository) FindTeamByID(teamID string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	var team models.Team
	err = r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) FindTeamByLeaderID(leaderID string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(leaderID)
	if err != nil {
		return nil, err
	}

	var team models.Team
	err = r.collection.FindOne(ctx, bson.M{"leader_id": id}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) FindTeamByMemberID(userID string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var team models.Team
	err = r.collection.FindOne(ctx, bson.M{"member_ids": id}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) FindTeamByInviteCode(code string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var team models.Team
	err := r.collection.FindOne(ctx, bson.M{"invite_code": code}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) FindTeamByName(name string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var team models.Team
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) UpdateTeam(team *models.Team) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	team.UpdatedAt = time.Now()
	filter := bson.M{"_id": team.ID}
	_, err := r.collection.ReplaceOne(ctx, filter, team)
	return err
}

func (r *TeamRepository) DeleteTeam(teamID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *TeamRepository) AddMemberToTeam(teamID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": teamObjID}
	update := bson.M{
		"$addToSet": bson.M{"member_ids": userObjID},
		"$set":      bson.M{"updated_at": time.Now()},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *TeamRepository) RemoveMemberFromTeam(teamID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": teamObjID}
	update := bson.M{
		"$pull": bson.M{"member_ids": userObjID},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *TeamRepository) UpdateTeamScore(teamID string, points int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": teamObjID}
	update := bson.M{
		"$inc": bson.M{"score": points},
		"$set": bson.M{"updated_at": time.Now()},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *TeamRepository) GetAllTeamsWithScores() ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err = cursor.All(ctx, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

// CountTeams returns the total number of teams
func (r *TeamRepository) CountTeams() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *TeamRepository) GetTeamMemberCount(teamID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return 0, err
	}

	var team models.Team
	err = r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&team)
	if err != nil {
		return 0, err
	}
	return len(team.MemberIDs), nil
}

// GetAllTeams returns all teams (for admin)
func (r *TeamRepository) GetAllTeams() ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err = cursor.All(ctx, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

// UpdateTeamFields updates specific fields of a team
func (r *TeamRepository) UpdateTeamFields(teamID primitive.ObjectID, fields bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fields["updated_at"] = time.Now()
	_, err := r.collection.UpdateByID(ctx, teamID, bson.M{"$set": fields})
	return err
}

// AdminDeleteTeam deletes a team by admin (bypasses leader check)
func (r *TeamRepository) AdminDeleteTeam(teamID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": teamID})
	return err
}

// AdminUpdateTeamLeader updates the team leader (admin only)
func (r *TeamRepository) AdminUpdateTeamLeader(teamID, newLeaderID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"leader_id":  newLeaderID,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateByID(ctx, teamID, update)
	return err
}

// GetRecentTeams returns teams created within the specified duration
func (r *TeamRepository) GetRecentTeams(since time.Time) ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"created_at": bson.M{"$gte": since}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err = cursor.All(ctx, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}
