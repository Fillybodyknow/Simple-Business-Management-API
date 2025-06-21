package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderHandle struct {
	OrderCollection    *mongo.Collection
	CustomerCollection *mongo.Collection
	ProductCollection  *mongo.Collection
}

type OrderItemRequest struct {
	ProductID string `json:"product_id" form:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" form:"quantity" binding:"required,min=1"`
}

type OrderRequest struct {
	Items            []OrderItemRequest `json:"items" form:"items" binding:"required,dive"`
	CustomerFullName string             `json:"customer_fullname" form:"customer_fullname" binding:"required"`
	CustomerEmail    string             `json:"customer_email" form:"customer_email" binding:"required,email"`
	CustomerPhone    string             `json:"customer_phone" form:"customer_phone" binding:"required"`
	CustomerAddress  string             `json:"customer_address" form:"customer_address" binding:"required"`
}

func NewOrderHandle(orderCol *mongo.Collection, customerCol *mongo.Collection, productCol *mongo.Collection) *OrderHandle {
	return &OrderHandle{OrderCollection: orderCol, CustomerCollection: customerCol, ProductCollection: productCol}
}

func (h *OrderHandle) CreatePublicOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input OrderRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔍 ดึงหรือสร้างลูกค้า
	var customer models.Customer
	err := h.CustomerCollection.FindOne(ctx, bson.M{"email": input.CustomerEmail}).Decode(&customer)
	if err == mongo.ErrNoDocuments {
		customer = models.Customer{
			FullName:  input.CustomerFullName,
			Email:     input.CustomerEmail,
			Phone:     input.CustomerPhone,
			Address:   input.CustomerAddress,
			CreatedAt: time.Now(),
		}
		result, err := h.CustomerCollection.InsertOne(ctx, customer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}
		customer.ID = result.InsertedID.(primitive.ObjectID)
	}

	var orderItems []models.OrderItem
	var totalAmount float64

	for _, item := range input.Items {
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID: " + item.ProductID})
			return
		}

		var product models.Product
		err = h.ProductCollection.FindOne(ctx, bson.M{"_id": productID, "is_active": true}).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found: " + item.ProductID})
			return
		}

		if item.Quantity > product.Stock {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Insufficient stock for %s", product.Name)})
			return
		}

		orderItems = append(orderItems, models.OrderItem{
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		})

		totalAmount += float64(item.Quantity) * product.Price
	}

	order := models.Order{
		CustomerID:  customer.ID,
		Status:      "Pending",
		TotalAmount: totalAmount,
		Items:       orderItems,
		CreatedAt:   time.Now(),
	}

	_, err = h.OrderCollection.InsertOne(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// 📦 หัก stock ของทุกสินค้า
	for _, item := range input.Items {
		productID, _ := primitive.ObjectIDFromHex(item.ProductID)
		_, err := h.ProductCollection.UpdateOne(ctx,
			bson.M{"_id": productID},
			bson.M{"$inc": bson.M{"stock": -item.Quantity}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order placed successfully"})
}
