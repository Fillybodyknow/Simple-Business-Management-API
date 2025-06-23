package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/models"
	"github.com/simple-business-management-api/go-backend-api/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ProductRepo repositories.ProductRepositoryInterface
}

func NewProductHandle(repo repositories.ProductRepositoryInterface) *ProductHandle {
	return &ProductHandle{ProductRepo: repo}
}

func (h *ProductHandle) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var products []models.Product

	products, err := h.ProductRepo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	exist, _ := h.ProductRepo.ExistsBySKU(ctx, input.SKU)
	if exist {
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
		return
	}

	err = h.ProductRepo.Insert(ctx, &product)
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

	userIdVar, _ := c.Get("userId")
	userIDStr, ok := userIdVar.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
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

	if exists, _ := h.ProductRepo.ExistsBySKU(ctx, input.SKU); exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists"})
		return
	}

	updateFields := bson.M{
		"name":      input.ProductName,
		"sku":       input.SKU,
		"price":     input.Price,
		"stock":     input.Stock,
		"is_active": input.IsActive,
	}

	result, err := h.ProductRepo.Update(ctx, productID, updateFields, roleVar.(string), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func (h *ProductHandle) DeleteProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roleVar, _ := c.Get("role")
	userIdVar, _ := c.Get("userId")
	userIDStr, ok := userIdVar.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
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

	result, err := h.ProductRepo.Delete(ctx, productID, roleVar.(string), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
