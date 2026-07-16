package helpers

import (
    "context"
    "fmt"
    "time"
    redis "github.com/redis/go-redis/v9"
)

func CheckRateLimit(ctx context.Context, rdb *redis.Client, userID uint64, limit int64, window time.Duration) (bool, error) {
    countKey := fmt.Sprintf("rate_limit:%d", userID)
    banKey := fmt.Sprintf("ban:%d", userID)

    isBanned, _ := rdb.Exists(ctx, banKey).Result()
    if isBanned > 0 {
        return false, nil 
    }

    count, err := rdb.Incr(ctx, countKey).Result()
    if err != nil {
        return false, err
    }

    if count == 1 {
        rdb.Expire(ctx, countKey, window)
    }

    if count > limit {
        rdb.Set(ctx, banKey, "true", 30 * time.Second)
        rdb.Del(ctx, countKey) 
        return false, nil
    }

    return true, nil
}