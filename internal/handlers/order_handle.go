package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"github.com/simple-business-management-api/go-backend-api/internal/pkg/utility"
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

type UpdateOrderRequest struct {
	Status         string `json:"status" form:"status" binding:"required"`
	Note           string `json:"note" form:"note"`
	TrackingNumber string `json:"tracking_number" form:"tracking_number"`
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

func (h *OrderHandle) CreateOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input OrderRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdVar, _ := c.Get("userId")
	Create_by, _ := primitive.ObjectIDFromHex(userIdVar.(string))
	if Create_by == primitive.NilObjectID {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	RoleVar, _ := c.Get("role")
	if RoleVar != "Admin" && RoleVar != "Staff" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only staff/admin can create orders"})
		return
	}

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
		CustomerID:      customer.ID,
		CreatedBy:       Create_by,
		Status:          "Pending",
		TotalAmount:     totalAmount,
		Items:           orderItems,
		CreatedAt:       time.Now(),
		Tracking_number: utility.GenerateTrackingNumber(),
		Note:            "อยู่ระหว่างดําเนินการ",
	}

	_, err = h.OrderCollection.InsertOne(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

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

func (h *OrderHandle) GetOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userIDVar, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID missing in context"})
		return
	}

	userIDStr, ok := userIDVar.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userIDNull, err := primitive.ObjectIDFromHex("000000000000000000000000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	roleVar, _ := c.Get("role")

	filter := bson.M{}
	if roleVar == "Staff" {
		filter = bson.M{
			"$or": []bson.M{
				{"created_by": userID},
				{"created_by": userIDNull},
			},
		}
	} else if roleVar != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	cursor, err := h.OrderCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandle) UpdateOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input UpdateOrderRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleVar, _ := c.Get("role")
	if roleVar != "Admin" && roleVar != "Staff" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only staff/admin can update orders"})
		return
	}

	orderID := c.Query("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing order ID"})
		return
	}

	orderIDHex, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	err = h.OrderCollection.FindOne(ctx, bson.M{"_id": orderIDHex}).Decode(&order)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if order.Status == "Cancelled" || order.Status == "Completed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update completed or cancelled order"})
		return
	}

	update := bson.M{
		"status": input.Status,
	}
	if input.Note != "" {
		update["note"] = input.Note
	}
	if input.TrackingNumber != "" {
		update["tracking_number"] = input.TrackingNumber
	}

	res, err := h.OrderCollection.UpdateOne(ctx, bson.M{"_id": orderIDHex}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}
