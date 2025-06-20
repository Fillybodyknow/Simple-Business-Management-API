package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
