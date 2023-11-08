package models

import "github.com/dgrijalva/jwt-go"

type LoginUserModel struct {
	UserId string `json:"userId"`
	Password string `json:"password"`
}

type JWTClaims struct {
	UserId string `json:"userId"`
	jwt.StandardClaims
}
