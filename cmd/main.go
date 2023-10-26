package main

import (
	"CloudKeep/handlers"
	"CloudKeep/initialise"
	"context"
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
		handlers.RegisterNewAccount(c, ctx, redisClient)
	})    
    router.POST("/api/register/verify-otp", func(c *gin.Context) {
		handlers.VerifyRegistrationOTP(c, ctx, redisClient, db)
	})    
    router.Run(os.Getenv("PORT"))

}