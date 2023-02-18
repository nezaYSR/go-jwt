package main

import (
	"os"

	routes "gitlab.com/nezaysr/golang-jwt/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	secret := os.Getenv("SECRET_KEY")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Use the cookie store for sessions
	store := cookie.NewStore([]byte(secret))
	router.Use(sessions.Sessions("session", store))

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": "Access granted for api-1",
		})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": "Access granted for api-2",
		})
	})

	router.Run(":" + port)
}
