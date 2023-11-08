package user_login_utils

import (
	"CloudKeep/models"
	"database/sql"
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