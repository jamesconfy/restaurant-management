package models

import "github.com/golang-jwt/jwt"

type SignedDetails struct {
	Email      string `json:"email"`
	First_Name string `json:"first_name"`
	Last_Name  string `json:"last_name"`
	User_ID    string `json:"user_id"`
	jwt.StandardClaims
}
