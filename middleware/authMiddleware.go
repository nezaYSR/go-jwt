package middleware

import (
	"errors"
	"fmt"
	"net/http"

	helper "gitlab.com/nezaysr/golang-jwt/helpers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)
		c.Next()

	}
}

func AuthenticateFromSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		tokenFound := session.Get("session")
		if tokenFound == nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "session not found"},
			)
			return
		}

		claims, err := helper.ValidateToken(tokenFound.(string))
		if err != "" || claims == nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "invalid token"},
			)
			return
		}

		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)
		c.Next()
	}
}

func InvalidateToken(token string) error {
	if token == "" {
		return errors.New("token not found")
	}
	return nil
}
