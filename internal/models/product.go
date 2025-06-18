package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	SKU       string             `bson:"sku"`
	Price     float64            `bson:"price"`
	Stock     int                `bson:"stock"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
}

type OrderItem struct {
	ProductID primitive.ObjectID `bson:"product_id"`
	Quantity  int                `bson:"quantity"`
	UnitPrice float64            `bson:"unit_price"`
}

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID  primitive.ObjectID `bson:"customer_id"`
	CreatedBy   primitive.ObjectID `bson:"created_by"`
	Status      string             `bson:"status"` // "pending", "paid", "shipped"
	TotalAmount float64            `bson:"total_amount"`
	Items       []OrderItem        `bson:"items"`
	CreatedAt   time.Time          `bson:"created_at"`
}
