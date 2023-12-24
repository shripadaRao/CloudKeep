package initialise

import (
	"CloudKeep/models"
	"CloudKeep/utils/ratelimitingutils"
	"CloudKeep/utils/redis_utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func InitializePostgres() (*sql.DB, error) {
    if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file while sending email: %v", err)
	}
    connStr := "user=" + os.Getenv("POSTGRES_USER") + " password=" + os.Getenv("POSTGRES_PASSWORD") +" dbname=" + os.Getenv("POSTGRES_DB") + " sslmode=disable" + " host=postgres" + " port=5432"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Printf("Error in postgres init: %v", err)
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        fmt.Printf("error in test ping: %v", err)
        return nil, err
    }

    fmt.Println("PostgreSQL initialized and connected.")
    return db, nil
}

func InitializeRedis() (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     "redis:6379", 
        Password: "",           
        DB:       0,            
    })

    _, err := client.Ping(context.Background()).Result()
    if err != nil {
        fmt.Printf("Error connecting to Redis: %v", err)
        return nil, err
    }

    fmt.Println("Redis initialized and connected.")
    return client, nil
}

func RateLimitTokens(redisClient *redis.Client, ctx context.Context) error {

    tokenBucketObj := models.TokenBucket{
        AvailableTokens: ratelimitingutils.GlobalRateLimit,
        LastRefilled: time.Now(),
    }

    return redis_utils.RedisSetData(ctx, redisClient, "rate_limit:global", tokenBucketObj, time.Hour*24)
}
