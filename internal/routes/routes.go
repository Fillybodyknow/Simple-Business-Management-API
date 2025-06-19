package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/simple-business-management-api/go-backend-api/internal/handlers"
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
	AuthHandle := handlers.NewAuthHandle(UserCollection)
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", AuthHandle.Register)
		}
	}

	return r
}
