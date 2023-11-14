package main

import (
	"CloudKeep/handlers"
	"CloudKeep/initialise"
	"CloudKeep/utils/user_login_utils"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func corsMiddleware(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(http.StatusNoContent)
        return
    }
    c.Next()
}

func main() {
    //init postgres and redis
    db, _ := initialise.InitializePostgres()
    defer db.Close()

    redisClient, _ := initialise.InitializeRedis()
    defer redisClient.Close()

    ctx := context.Background()

    //routes
    router := gin.Default()
    router.Use(corsMiddleware)

    router.GET("/hello", func(c *gin.Context) {
        c.String(http.StatusOK, "Hello World")
    })

    router.GET("/api/validate-jwt", func(c *gin.Context) {
        handlers.ValidateJWT_API(c)
    })
    
	router.POST("/api/register/send-email-otp", func(c *gin.Context) {
		handlers.SendRegistrationEmail(c, ctx, redisClient)
	})    
    router.POST("/api/register/verify-otp", func(c *gin.Context) {
		handlers.VerifyRegistrationOTP(c, ctx, redisClient)
	})    
    router.POST("/api/register/create-user", func(c *gin.Context) {
        handlers.CreateUser(c, ctx, redisClient, db)
    })
    
    router.POST("/api/login/userid-password", func(c *gin.Context){
        handlers.LoginUserByUserIdPassword(c, ctx, db)
    })

    // ------- auth middleware ---------- //
	uploadGroup := router.Group("/api/upload")
	{
		uploadGroup.POST("/init", authMiddleware, func(c *gin.Context) {
			handlers.InitializeUploadProcess(c, db)
		})

		uploadGroup.POST("/chunk", authMiddleware, func(c *gin.Context) {
			handlers.UploadChunk(c, db)
		})

		uploadGroup.POST("", authMiddleware, func(c *gin.Context) {
			handlers.MergeChunks(c, db)
		})
	}

    err := router.Run(os.Getenv("PORT"))
    if err != nil {
        log.Fatalf("Error in running server: %v", err)
    }
}

func authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		c.Abort()
		return
	}

	// Parse the Authorization header using ParseAuthHeader function
	tokenString, err := user_login_utils.ParseAuthHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	claims, err := user_login_utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT"})
		c.Abort()
		return
	}

	// Set the validated claims in the context for use in the route handler
	c.Set("claims", claims)
	c.Next()
}
