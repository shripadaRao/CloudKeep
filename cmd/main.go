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

func main() {
    //init postgres and redis
    db, _ := initialise.InitializePostgres()
    defer db.Close()

    redisClient, _ := initialise.InitializeRedis()
    defer redisClient.Close()

    ctx := context.Background()

    //routes
    router := gin.Default()

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

    err := router.Run(os.Getenv("PORT"))
    if err != nil {
        log.Fatalf("Error in running server: %v", err)
    }
}