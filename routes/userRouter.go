package routes

import (
	controller "gitlab.com/nezaysr/golang-jwt/controllers"
	"gitlab.com/nezaysr/golang-jwt/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.AuthenticateFromSession())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
