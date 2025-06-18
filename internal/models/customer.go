package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	FullName  string             `bson:"full_name"`
	Email     string             `bson:"email"`
	Phone     string             `bson:"phone"`
	Address   string             `bson:"address"`
	CreatedAt time.Time          `bson:"created_at"`
}
