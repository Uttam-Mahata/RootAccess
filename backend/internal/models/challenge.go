package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Challenge struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	Points      int                `bson:"points" json:"points"`
	Flag        string             `bson:"flag" json:"-"` // Hidden from user
	Files       []string           `bson:"files" json:"files"`
}
