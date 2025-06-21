package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/handlers"
	"github.com/simple-business-management-api/go-backend-api/internal/middleware"
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
	ProductHandle := handlers.NewProductHandle(ProductCollection)
	OrderHandle := handlers.NewOrderHandle((OrderCollection), (CustomerCollection), (ProductCollection))
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", AuthHandle.Register)
			auth.POST("/login", AuthHandle.Login)
		}
		product := api.Group("/product")
		{
			product.GET("/", ProductHandle.GetProducts)
		}
		productMiddleware := api.Group("/product")
		productMiddleware.Use(middleware.AuthMiddleware())
		{
			productMiddleware.POST("/", ProductHandle.CreateProduct)
			productMiddleware.PUT("", ProductHandle.UpdateProduct)
			productMiddleware.DELETE("", ProductHandle.DeleteProduct)
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
