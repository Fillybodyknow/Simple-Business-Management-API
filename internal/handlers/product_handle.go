package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	SKU         string  `json:"sku" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
}
type ProductHandle struct {
	ProductCollection *mongo.Collection
}

func NewProductHandle(productCol *mongo.Collection) *ProductHandle {
	return &ProductHandle{ProductCollection: productCol}
}

func (h *ProductHandle) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cursor, err := h.ProductCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	}
	defer cursor.Close(ctx)

	var products []models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode product"})
		}
		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(products),
		"products": products,
	})

}

func (h *ProductHandle) CreateProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roleVar, _ := c.Get("role")
	if roleVar != "Admin" && roleVar != "Staff" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only staff/admin can create products"})
		return
	}

	userIdVar, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	createBy, err := primitive.ObjectIDFromHex(userIdVar.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var input ProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:      input.ProductName,
		CreatedBy: createBy,
		SKU:       input.SKU,
		Price:     input.Price,
		Stock:     input.Stock,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	_, err = h.ProductCollection.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully"})
}
