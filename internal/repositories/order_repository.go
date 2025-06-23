package repositories

import (
	"context"
	"fmt"

	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepositoryInterface interface {
	FindAll(ctx context.Context, userID primitive.ObjectID, role string) ([]models.Order, error)
	Insert(ctx context.Context, order *models.Order, role string) error
	FindByID(ctx context.Context, id primitive.ObjectID, role string) (*models.Order, error)
	Update(ctx context.Context, id primitive.ObjectID, fields bson.M, role string) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, role string) (*mongo.DeleteResult, error)
}

type OrderRepository struct {
	Collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) *OrderRepository {
	return &OrderRepository{Collection: collection}
}

func (r *OrderRepository) FindAll(ctx context.Context, userID primitive.ObjectID, role string) ([]models.Order, error) {
	var filter bson.M

	userIDNull, err := primitive.ObjectIDFromHex("000000000000000000000000")
	if err != nil {
		return nil, err
	}

	switch role {
	case "Admin":
		filter = bson.M{}
	case "Staff":
		filter = bson.M{"$or": []bson.M{{"created_by": userID}, {"created_by": userIDNull}}}
	default:
		return nil, fmt.Errorf("unauthorized role")
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *OrderRepository) Insert(ctx context.Context, order *models.Order, role string) error {
	if role != "Admin" && role != "Staff" {
		return fmt.Errorf("unauthorized role")
	}
	_, err := r.Collection.InsertOne(ctx, order)
	return err
}

func (r *OrderRepository) FindByID(ctx context.Context, id primitive.ObjectID, role string) (*models.Order, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	var order models.Order
	if err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Update(ctx context.Context, id primitive.ObjectID, fields bson.M, role string) (*mongo.UpdateResult, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	result, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
	return result, err
}

func (r *OrderRepository) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, role string) (*mongo.DeleteResult, error) {

	filter := bson.M{"_id": id}
	if role == "Staff" {
		filter["$or"] = []bson.M{
			{"created_by": userID},
			{"created_by": bson.M{"$exists": false}},
		}
	}
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	result, err := r.Collection.DeleteOne(ctx, filter)
	return result, err
}
