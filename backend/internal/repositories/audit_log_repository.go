package repositories

import (
	"context"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepository struct {
	collection *mongo.Collection
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{
		collection: database.DB.Collection("audit_logs"),
	}
}

func (r *AuditLogRepository) CreateLog(log *models.AuditLog) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *AuditLogRepository) GetLogs(limit int, page int) ([]models.AuditLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	skip := (page - 1) * limit

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditLogRepository) GetLogCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}
