package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedBy primitive.ObjectID `bson:"created_by"`
	Name      string             `bson:"name"`
	SKU       string             `bson:"sku"`
	Price     float64            `bson:"price"`
	Stock     int                `bson:"stock"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
}
