package helper

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jamesconfy/restaurant-management/database"
	"github.com/jamesconfy/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var secret_key string = os.Getenv("SECRET_KEY")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GenerateAllToken(email, first_name, last_name, userId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &models.SignedDetails{
		Email:      email,
		First_Name: first_name,
		Last_Name:  last_name,
		User_ID:    userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(60)).Unix(),
		},
	}

	refreshClaims := &models.SignedDetails{
		Email:      email,
		First_Name: first_name,
		Last_Name:  last_name,
		User_ID:    userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret_key))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret_key))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func UpdateToken(signedToken, signedRefreshToken, userId string) error {
	var updateObj primitive.D
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opts)
	if err != nil {
		log.Panic(err.Error())
	}

	return nil
}

func ValidateToken(signedToken string) (claims *models.SignedDetails, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&models.SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret_key), nil
		},
	)

	claims, ok := token.Claims.(*models.SignedDetails)
	if !ok {
		return nil, err
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, err
	}

	return claims, err

}
