package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gitlab.com/nezaysr/golang-jwt/database"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, user_id string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")

	err = nil
	if userType == "USER" && uid != user_id {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}
