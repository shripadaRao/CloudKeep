package handlers

import (
	"CloudKeep/utils/user_login_utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateJWT_API(c *gin.Context) {
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}

	tokenString, err := user_login_utils.ParseAuthHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	claims, err := user_login_utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "JWT is valid", "userId": claims.UserId})
}
