package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/handlers"
	"github.com/simple-business-management-api/go-backend-api/internal/middleware"
	"github.com/simple-business-management-api/go-backend-api/internal/repositories"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetRoutes(db *mongo.Client) *gin.Engine {
	r := gin.Default()

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	UserCollection := db.Database("Simple-Business-Management").Collection("users")
	ProductCollection := db.Database("Simple-Business-Management").Collection("products")
	OrderCollection := db.Database("Simple-Business-Management").Collection("orders")
	CustomerCollection := db.Database("Simple-Business-Management").Collection("customers")
	AuthHandle := handlers.NewAuthHandle(UserCollection)
	productRepo := repositories.NewProductRepository(ProductCollection)
	OrderHandle := handlers.NewOrderHandle((OrderCollection), (CustomerCollection), (ProductCollection))
	productHandler := handlers.NewProductHandle(productRepo)
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", AuthHandle.Register)
			auth.POST("/login", AuthHandle.Login)
		}
		product := api.Group("/product")
		{
			product.GET("/", productHandler.GetProducts)
		}
		productMiddleware := api.Group("/product")
		productMiddleware.Use(middleware.AuthMiddleware())
		{
			productMiddleware.POST("/", productHandler.CreateProduct)
			productMiddleware.PUT("", productHandler.UpdateProduct)
			productMiddleware.DELETE("", productHandler.DeleteProduct)
		}
		orderMiddleware := api.Group("/order")
		orderMiddleware.Use(middleware.AuthMiddleware())
		{
			orderMiddleware.POST("/", OrderHandle.CreateOrders)
			orderMiddleware.GET("/", OrderHandle.GetOrders)
			orderMiddleware.PUT("", OrderHandle.UpdateOrder)
			orderMiddleware.DELETE("", OrderHandle.DeleteOrder)
		}
	}
	PublicAPI := r.Group("/api/public")
	{
		Order := PublicAPI.Group("/order")
		{
			Order.POST("/", OrderHandle.CreatePublicOrder)
		}
	}

	return r
}
