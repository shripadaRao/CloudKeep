package redis_utils

import (
	"CloudKeep/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisSetData(ctx context.Context, r *redis.Client, key string, object interface{}, expiration time.Duration) error {
    jsonData, err := json.Marshal(object)
    if err != nil {
        fmt.Println("Error in parsing object in RedisSetData: ", err)
        return err
    }

    err = r.Set(ctx, key, jsonData, expiration).Err()
    if err != nil {
        return err
    }

    return nil
}

func RedisGetData(ctx context.Context, r *redis.Client, key string) (interface{}, error) {
    val, err := r.Get(ctx, key).Result()
    if err != nil {
        log.Println("Error fetching the key: ", err)
        if err == redis.Nil {
            // key not found, most probably
            log.Println("Key not found: ", err)
            return nil, err
        }
        return nil, err
    }
    return val, nil
}

func ParseTokenBucketString(data string) (*models.TokenBucket, error) {
    var bucket models.TokenBucket
    err := json.Unmarshal([]byte(data), &bucket)
    if err != nil {
        return nil, err
    }
    return &bucket, nil
}

