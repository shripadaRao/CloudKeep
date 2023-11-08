package handlers

import (
	"CloudKeep/models"
	"CloudKeep/utils/user_login_utils"
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)


func LoginUserByUserIdPassword(c *gin.Context, ctx context.Context, db *sql.DB) {
	var userLoginCreds models.LoginUserModel
	if err := c.ShouldBindJSON(&userLoginCreds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dbUser *models.User
	dbUser, err := user_login_utils.GetUserPasswordByUserId(db, userLoginCreds.UserId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Was unable to fetch user details from UserId", "error":err.Error()})
		return
	}

	if !user_login_utils.VerifyPassword(user_login_utils.ConstructUserLoginPassword(userLoginCreds.Password, dbUser.Salt), dbUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	tokenString, err := user_login_utils.GenerateJWT(userLoginCreds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating JWT token", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

//todo later
func LoginUserByEmailOTP(c *gin.Context, ctx context.Context, db *sql.DB){
	return
}