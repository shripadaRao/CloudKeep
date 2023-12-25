package main

import (
	"CloudKeep/handlers"
	"CloudKeep/initialise"
	"CloudKeep/models"
	"CloudKeep/utils/ratelimitingutils"
	"CloudKeep/utils/user_login_utils"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	initialise.RateLimitTokens(redisClient, ctx)

    //routes
    router := gin.Default()
    router.Use(corsMiddleware)

    router.GET("/hello", rateLimitMiddleware(redisClient), func(c *gin.Context) {
        c.String(http.StatusOK, "Hello World")
    })

    router.GET("/api/validate-jwt", rateLimitMiddleware(redisClient), func(c *gin.Context) {
        handlers.ValidateJWT_API(c)
    })
    
	router.POST("/api/register/send-email-otp", rateLimitMiddleware(redisClient), func(c *gin.Context) {
		handlers.SendRegistrationEmail(c, ctx, redisClient)
	})    
    router.POST("/api/register/verify-otp", rateLimitMiddleware(redisClient), func(c *gin.Context) {
		handlers.VerifyRegistrationOTP(c, ctx, redisClient)
	})    
    router.POST("/api/register/create-user", rateLimitMiddleware(redisClient), func(c *gin.Context) {
        handlers.CreateUser(c, ctx, redisClient, db)
    })
    
    router.POST("/api/login/userid-password", rateLimitMiddleware(redisClient), func(c *gin.Context){
        handlers.LoginUserByUserIdPassword(c, ctx, db)
    })

    // ------- auth middleware ---------- //
	uploadGroup := router.Group("/api/upload")
	{
		uploadGroup.Use(authMiddleware)

		uploadGroup.POST("/simple-upload", rateLimitMiddleware(redisClient), func(c *gin.Context) {
			handlers.SimpleTestUploadAPI(c, db)
		})

		uploadGroup.POST("/init", rateLimitMiddleware(redisClient), func(c *gin.Context) {
			handlers.InitializeUploadProcess(c, db)
		})

		uploadGroup.POST("/chunk", rateLimitMiddleware(redisClient), func(c *gin.Context) {
			handlers.UploadChunk(c, db)
		})

		uploadGroup.POST("/merge", rateLimitMiddleware(redisClient), func(c *gin.Context) {
			handlers.MergeChunks(c, db)
		})
		uploadGroup.POST("/bucket", rateLimitMiddleware(redisClient), func(c *gin.Context) {
			handlers.StoreMergedFileS3(c, db)
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

func rateLimitMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		var userId string
		if exists {
			userId = claims.(*models.JWTClaims).UserId
		} else {
			userId = ""
		}

		apiPath := c.Request.URL.Path
		if !ratelimitingutils.IsAllowedByRateLimit(userId, apiPath, redisClient) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
