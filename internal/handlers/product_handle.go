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
	ProductName string  `json:"product_name" form:"product_name" binding:"required"`
	SKU         string  `json:"sku" form:"sku" binding:"required"`
	Price       float64 `json:"price" form:"price" binding:"required"`
	Stock       int     `json:"stock" form:"stock" binding:"required"`
}

type UpdateProductRequest struct {
	ProductName string  `json:"product_name" form:"product_name"`
	SKU         string  `json:"sku" form:"sku"`
	Price       float64 `json:"price" form:"price"`
	Stock       int     `json:"stock" form:"stock"`
	IsActive    bool    `json:"is_active" form:"is_active"`
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
	if err := c.ShouldBind(&input); err != nil {
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

func (h *ProductHandle) UpdateProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roleVar, _ := c.Get("role")
	if roleVar != "Admin" && roleVar != "Staff" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only staff/admin can update products"})
		return
	}

	productIDStr := c.Query("id")
	if productIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing product ID"})
		return
	}
	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var input = UpdateProductRequest{}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":      input.ProductName,
			"sku":       input.SKU,
			"price":     input.Price,
			"stock":     input.Stock,
			"is_active": input.IsActive,
		},
	}

	resulf, err := h.ProductCollection.UpdateOne(ctx, bson.M{"_id": productID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if resulf.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func (h *ProductHandle) DeleteProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roleVar, _ := c.Get("role")
	productIDStr := c.Query("id")
	if productIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing product ID"})
		return
	}

	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	switch roleVar {
	case "Staff":
		userIdVar, _ := c.Get("userId")
		userIDStr, ok := userIdVar.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		filter := bson.M{"_id": productID, "created_by": userID}
		result, err := h.ProductCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found or permission denied"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
		return

	case "Admin":
		result, err := h.ProductCollection.DeleteOne(ctx, bson.M{"_id": productID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
		return

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Only staff/admin can delete products"})
	}
}
