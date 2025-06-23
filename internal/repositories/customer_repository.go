package repositories

import (
	"context"
	"fmt"

	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomerRepositoryInterface interface {
	FindAll(ctx context.Context, role string) ([]models.Customer, error)
	Insert(ctx context.Context, customer *models.Customer, role string) (*mongo.InsertOneResult, error)
	FindByEmail(ctx context.Context, email string, role string) (*models.Customer, error)
	FindByID(ctx context.Context, id string, role string) (*models.Customer, error)
}

type CustomerRepository struct {
	Collection *mongo.Collection
}

func NewCustomerRepository(collection *mongo.Collection) *CustomerRepository {
	return &CustomerRepository{Collection: collection}
}

func (r *CustomerRepository) FindAll(ctx context.Context, role string) ([]models.Customer, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []models.Customer
	for cursor.Next(ctx) {
		var customer models.Customer
		if err := cursor.Decode(&customer); err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

func (r *CustomerRepository) Insert(ctx context.Context, customer *models.Customer, role string) (*mongo.InsertOneResult, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	result, err := r.Collection.InsertOne(ctx, customer)
	return result, err
}

func (r *CustomerRepository) FindByEmail(ctx context.Context, email string, role string) (*models.Customer, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	var customer models.Customer
	if err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string, role string) (*models.Customer, error) {
	if role != "Admin" && role != "Staff" {
		return nil, fmt.Errorf("unauthorized role")
	}
	var customer models.Customer
	if err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}
