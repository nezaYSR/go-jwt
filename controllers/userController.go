package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	helper "gitlab.com/nezaysr/golang-jwt/helpers"
	"gitlab.com/nezaysr/golang-jwt/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/nezaysr/golang-jwt/database"
	"gitlab.com/nezaysr/golang-jwt/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password incorrect")
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			// log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error ocured while checking for email"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error ocured while checking for phone number"})
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exist"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)

		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)

		// Store token in session cookie
		session := sessions.Default(c)
		session.Set("session", token)
		session.Save()

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientTokenFound := c.Request.Header.Get("token")
		err := middleware.InvalidateToken(clientTokenFound)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
			return
		}

		isTokenDeletedErr := helper.WipeoutAllFuckingField(c, clientTokenFound)

		if isTokenDeletedErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": isTokenDeletedErr})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func LogoutFromSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		clientTokenFound := session.Get("session")
		if clientTokenFound == "" || clientTokenFound == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "token not exist"})
			return
		}

		clientTokenFoundValue := clientTokenFound.(string)

		err := middleware.InvalidateToken(clientTokenFoundValue)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
			return
		}

		isTokenDeletedErr := helper.WipeoutAllFuckingField(c, clientTokenFoundValue)

		if isTokenDeletedErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": isTokenDeletedErr})
			return
		}

		session.Set("session", "") // this will mark the session as "written" and hopefully remove the username
		session.Clear()
		session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
		session.Save()

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := helper.CheckUserType(c, "ADMIN")
		if err != nil {
			// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Println("Error generating token:", err)
			return
		}
		_, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}

		groupStage := bson.D{
			{"$group", bson.D{
				{"_id", bson.D{
					{"_id", "null"},
				},
				},
				{
					"totalCount", bson.D{
						{"$sum", 1},
					},
				},
				{
					"data", bson.D{
						{"$push", "$$ROOT"},
					},
				},
			}},
		}

		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"totalCount", 1},
				{"userItem", bson.D{
					{"$slice", []interface{}{"$data", startIndex, recordPerPage}}},
				}},
			},
		}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}

		var allUsers []bson.M

		if err = result.All(c, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, user_id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": user_id}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}
