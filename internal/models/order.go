package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ProductID primitive.ObjectID `bson:"product_id"`
	Quantity  int                `bson:"quantity"`
	UnitPrice float64            `bson:"unit_price"`
}

type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID      primitive.ObjectID `bson:"customer_id"`
	CreatedBy       primitive.ObjectID `bson:"created_by"`
	Tracking_number string             `bson:"tracking_number"`
	Note            string             `bson:"note"`
	Status          string             `bson:"status"` // "pending", "paid", "shipped"
	TotalAmount     float64            `bson:"total_amount"`
	Items           []OrderItem        `bson:"items"`
	CreatedAt       time.Time          `bson:"created_at"`
}
