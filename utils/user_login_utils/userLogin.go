package user_login_utils

import (
	"CloudKeep/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your-secret-key")

func GenerateJWT(user models.LoginUserModel) (string, error){
	JWT_ExpirationTime := time.Now().Add(24 * time.Hour) 

	var claims models.JWTClaims

	claims.UserId = user.UserId
	claims.StandardClaims = jwt.StandardClaims{ExpiresAt:JWT_ExpirationTime.Unix()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}

func ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid JWT")
	}

	return claims, nil
}

func GetUserPasswordByUserId(db *sql.DB, userId string) (*models.User, error) {
	var dbUser models.User
	query := `SELECT userid, password, salt FROM user_table WHERE userid = $1`

	err := db.QueryRow(query, userId).Scan(&dbUser.UserID, &dbUser.Password, &dbUser.Salt)
	if err != nil {
		return nil, err
	}
	return &dbUser, nil
}

func ConstructUserLoginPassword(password string, salt string) string{
	saltedPassword := password + salt
	return saltedPassword
}

func VerifyPassword(userInputPassword string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(userInputPassword))
	return err == nil
}