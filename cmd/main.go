package main

import (
	"CloudKeep/handlers"
	"CloudKeep/initialise"
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
    router.POST("/api/upload/init", func(c *gin.Context) {
        handlers.InitializeUploadProcess(c, db)
    })
    router.POST("/api/upload/chunk", func(c *gin.Context) {
        handlers.UploadChunk(c, db)
    })

    err := router.Run(os.Getenv("PORT"))
    if err != nil {
        log.Fatalf("Error in running server: %v", err)
    }
}