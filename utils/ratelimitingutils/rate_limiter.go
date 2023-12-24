package ratelimitingutils

import (
	"CloudKeep/models"
	"CloudKeep/utils/redis_utils"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
    GlobalRateLimit   = 100   // Global limit per minute
    rateLimitWindow   = 60 * time.Second
	defaultApiPath = "/"
)

type ApiLimit struct {
    Limit      int
    RefillRate float64
}

var apiLimits = map[string]ApiLimit{
    "/hello":                       {Limit: 100, RefillRate: 1},
    "/api/validate-jwt":            {Limit: 50, RefillRate: 0.5},
    "/api/register/send-email-otp": {Limit: 10, RefillRate: 0.1},
    "/api/register/verify-otp":     {Limit: 5, RefillRate: 0.1},
    "/api/register/create-user":    {Limit: 1, RefillRate: 0.05},
    "/api/login/userid-password":   {Limit: 10, RefillRate: 0.5},
    "/api/upload/*":                {Limit: 5, RefillRate: 0.1},
    "/":                            {Limit: 10, RefillRate: 0.05},
}


func IsAllowedByRateLimit(userId string, apiPath string, redisClient *redis.Client) bool {
	if !isAllowedByGlobalRateLimit(apiPath, redisClient) {
		return false
	}

	// future work
	// if !isRateLimitedForUser(userId, apiPath, redisClient) {
	// 	return false
	// }

	return true
}

func isAllowedByGlobalRateLimit(apiPath string, redisClient *redis.Client) bool {
	if !successfullyConsumedGlobalTokens(apiPath, redisClient) {
		return false
	}
	return true
}


func successfullyConsumedGlobalTokens(apiPath string, redisClient *redis.Client) bool {
	key := fmt.Sprintf("rate_limit:global")
	return successfullyConsumedTokens(key, GlobalRateLimit, redisClient)
}


func getTokenBucketObj(ctx context.Context, key string, redisClient *redis.Client) (models.TokenBucket, error){

	tokenBucketObjString, err := redis_utils.RedisGetData(ctx, redisClient, key)

	var tokenBucketObjPtr *models.TokenBucket
	tokenBucketObjPtr, err = redis_utils.ParseTokenBucketString(tokenBucketObjString.(string))

	if err != nil {
		fmt.Println("Error in parsing rate limit token value while consuming tokens", err)
		return models.TokenBucket{}, err
	}
	tokenBucketObj := *tokenBucketObjPtr

	return tokenBucketObj, nil
}

func successfullyConsumedTokens(key string, limit int, redisClient *redis.Client) bool {
	ctx := context.TODO()

	var tokenBucketObj models.TokenBucket
	tokenBucketObj, err := getTokenBucketObj(ctx, key, redisClient)

	tokenBucketObj.AvailableTokens--

	if tokenBucketObj.AvailableTokens > 0 {
		err = redis_utils.RedisSetData(ctx, redisClient, key, tokenBucketObj, time.Hour*24)

		if err != nil {
			fmt.Println("Error in decrementing and setting rate limit token value while consuming tokens", err)
			return false
		}
		return true
	} 

	if refillTokenMechanism(key, tokenBucketObj ,limit, redisClient) {
		return true
	}

	return false
}

func refillTokens(key string, tokenBucketObj models.TokenBucket, maxTokenLimit int, redisClient *redis.Client) error{
	ctx := context.TODO()
	tokenBucketObj.AvailableTokens = maxTokenLimit
	tokenBucketObj.LastRefilled = time.Now()
	err := redis_utils.RedisSetData(ctx, redisClient, key, tokenBucketObj, time.Hour)
	return err
}

func refillTokenMechanism(key string, tokenBucketObj models.TokenBucket, limit int, redisClient *redis.Client) bool{
		// refill condition
		if time.Since(tokenBucketObj.LastRefilled) > rateLimitWindow {
			err := refillTokens(key , tokenBucketObj, limit, redisClient)
			if err != nil {
				fmt.Println("Error in refilling tokens", err)
			}
			return true
		} 
		return false
}

// future work
func getApiLimit(apiPath string) (ApiLimit, bool) {
    if limitInfo, exists := apiLimits[apiPath]; exists {
        return limitInfo, true
    }
    return ApiLimit{}, false
}

func consumeUserTokens(userId, apiPath string, limit int, redisClient *redis.Client) bool {
	key := fmt.Sprintf("rate_limit:user:%s", userId)
	return successfullyConsumedTokens(key, limit, redisClient)
}


func isRateLimitedForUser(userId, apiPath string, redisClient *redis.Client) bool {
	if userId != "" {
		apiLimit, exists := getApiLimit(apiPath)
		if !exists {
			apiLimit = ApiLimit{Limit:100, RefillRate: 0.5}
		}
		if !consumeUserTokens(userId, apiPath, apiLimit.Limit, redisClient) {
			return false
		}
	}

	return true
}