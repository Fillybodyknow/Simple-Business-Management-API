package repositories

import (
	"context"
	"fmt"

	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepositoryInterface interface {
	FindAll(ctx context.Context) ([]models.Product, error)
	FindByID(ctx context.Context, id primitive.ObjectID, is_active bool) (*models.Product, error)
	Insert(ctx context.Context, product *models.Product) error
	UpdateStock(ctx context.Context, id primitive.ObjectID, NewStock int) error
	Update(ctx context.Context, id primitive.ObjectID, filter bson.M, role string, userID primitive.ObjectID) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, productID primitive.ObjectID, role string, userID primitive.ObjectID) (*mongo.DeleteResult, error)
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
}

type ProductRepository struct {
	Collection *mongo.Collection
}

func NewProductRepository(collection *mongo.Collection) *ProductRepository {
	return &ProductRepository{Collection: collection}
}

func (r *ProductRepository) FindAll(ctx context.Context) ([]models.Product, error) {
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var Products []models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		Products = append(Products, product)
	}
	return Products, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id primitive.ObjectID, is_active bool) (*models.Product, error) {
	var product models.Product
	if err := r.Collection.FindOne(ctx, bson.M{"_id": id, "is_active": is_active}).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Insert(ctx context.Context, product *models.Product) error {
	_, err := r.Collection.InsertOne(ctx, product)
	return err
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id primitive.ObjectID, NewStock int) error {
	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"stock": NewStock}})
	return err
}

func (r *ProductRepository) Update(ctx context.Context, productID primitive.ObjectID, fields bson.M, role string, userID primitive.ObjectID) (*mongo.UpdateResult, error) {

	var filter bson.M

	switch role {
	case "Admin":
		filter = bson.M{"_id": productID}
	case "Staff":
		filter = bson.M{"_id": productID, "created_by": userID}
	default:
		return nil, fmt.Errorf("unauthorized role")
	}

	result, err := r.Collection.UpdateOne(ctx, filter, bson.M{"$set": fields})
	return result, err
}

func (r *ProductRepository) Delete(ctx context.Context, productID primitive.ObjectID, role string, userID primitive.ObjectID) (*mongo.DeleteResult, error) {

	var filter bson.M

	switch role {
	case "Admin":
		filter = bson.M{"_id": productID}
	case "Staff":
		filter = bson.M{"_id": productID, "created_by": userID}
	default:
		return nil, fmt.Errorf("unauthorized role")
	}

	result, err := r.Collection.DeleteOne(ctx, filter)
	return result, err
}

func (r *ProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	count, err := r.Collection.CountDocuments(ctx, bson.M{"sku": sku})
	return count > 0, err
}
