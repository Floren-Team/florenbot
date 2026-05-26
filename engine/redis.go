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
	// 1. Берем хост из переменных окружения (в docker-compose это "redis")
	// Если переменная пустая, используем "redis:6379" как дефолт для Docker
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db, _ := strconv.Atoi(dbStr)

	log.Printf("🔍 [DEBUG] Попытка подключения к Redis по адресу: %s", addr)

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		Network:  "tcp4", // ПРИНУДИТЕЛЬНО используем IPv4, чтобы убрать ошибку [::1]
	})

	// Проверка соединения
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Предупреждение: Redis недоступен (%v). Работаем без кэша.", err)
	} else {
		log.Println("✅ Redis успешно подключен")
	}
}

// GetBalance ищет баланс в Redis.
func GetBalance(id int64, username string) (int, error) {
	key := fmt.Sprintf("user:%d:balance", id)
	
	// Если RDB не инициализирован или ошибка, идем сразу в SQL
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

// ChangeBalance обновляет баланс.
func ChangeBalance(id int64, amount int) error {
	err := UpdateBalanceSQL(id, amount)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("user:%d:balance", id)
	// Если Redis недоступен, IncrBy вернет ошибку, мы просто пропустим кэш
	_, err = RDB.IncrBy(ctx, key, int64(amount)).Result()
	if err != nil {
		RDB.Del(ctx, key)
	}
	return nil
}