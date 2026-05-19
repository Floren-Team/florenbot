package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	
	db, _ := strconv.Atoi(dbStr)

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("Предупреждение: Redis недоступен (%v). Работа замедлится.", err)
	}
}

// GetBalance ищет баланс в Redis. Если пусто — идет в SQLite (где создается юзер) и кэширует.
func GetBalance(id int64, username string) (int, error) {
	key := fmt.Sprintf("user:%d:balance", id)
	
	val, err := RDB.Get(ctx, key).Result()
	if err == nil {
		balance, _ := strconv.Atoi(val)
		return balance, nil
	}

	balance, err := GetUserBalanceSQL(id, username)
	if err != nil {
		return 0, err
	}

	RDB.Set(ctx, key, balance, 24*time.Hour)
	return balance, nil
}

// ChangeBalance обновляет баланс и в SQLite, и в кэше Redis
func ChangeBalance(id int64, amount int) error {
	err := UpdateBalanceSQL(id, amount)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:balance", id)
	_, err = RDB.IncrBy(ctx, key, int64(amount)).Result()
	if err != nil {
		RDB.Del(ctx, key)
	}
	return nil
}