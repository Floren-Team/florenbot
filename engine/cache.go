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

// InitCache инициализирует кэширование в зависимости от настроек
func InitCache() {
	engine := getEnv("CACHE_ENGINE", "local")

	if engine == "redis" {
		addr := getEnv("REDIS_ADDR", "redis:6379")
		password := os.Getenv("REDIS_PASSWORD")
		db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

		RDB = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
			Network:  "tcp4",
		})

		if err := RDB.Ping(ctx).Err(); err != nil {
			log.Printf("⚠️ Redis недоступен: %v. Работаем без кэша.", err)
		} else {
			log.Println("✅ Redis успешно подключен")
		}
	} else {
		log.Println("📁 Используется локальный кэш (папка cache)")
		if _, err := os.Stat("cache"); os.IsNotExist(err) {
			if err := os.Mkdir("cache", 0755); err != nil && !os.IsExist(err) {
				log.Printf("Помилка створення папки: %v", err)
			}
		}
	}
}

// SetCache сохраняет значение в выбранный движок
func SetCache(key string, value string, duration time.Duration) {
	if getEnv("CACHE_ENGINE", "local") == "redis" {
		RDB.Set(ctx, key, value, duration)
	} else {
		// Используем подчеркивания вместо двоеточий для имен файлов
		if err := os.WriteFile("cache/"+key, []byte(value), 0644); err != nil {
			log.Printf("Ошибки сохранения в кэш: %v", err)
		}
	}
}

// GetCache получает значение из выбранного движка
func GetCache(key string) (string, error) {
	if getEnv("CACHE_ENGINE", "local") == "redis" {
		return RDB.Get(ctx, key).Result()
	} else {
		data, err := os.ReadFile("cache/" + key)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

// ClearCache удаляет ключ из кэша
func ClearCache(key string) {
	if getEnv("CACHE_ENGINE", "local") == "redis" {
		RDB.Del(ctx, key)
	} else {
		os.Remove("cache/" + key)
	}
}

// GetBalance с использованием универсального кэша
func GetBalance(id uint64, username string) (int, error) {
	key := fmt.Sprintf("user_%d_balance", id)

	val, err := GetCache(key)
	if err == nil {
		balance, _ := strconv.Atoi(val)
		return balance, nil
	}

	balance, err := GetUserBalanceSQL(id, username)
	if err != nil {
		return 0, err
	}

	SetCache(key, strconv.Itoa(balance), 24*time.Hour)
	return balance, nil
}

// ChangeBalance обновляет баланс и инвалидирует кэш
func ChangeBalance(id uint64, amount int) error {
	err := UpdateBalanceSQL(id, amount)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("user_%d_balance", id)
	ClearCache(key) // Просто удаляем, чтобы при следующем запросе обновилось из БД
	return nil
}

func CloseRedis() {
	if RDB != nil {
		log.Println("💾 Закрытие подключения к Redis...")

		RDB.Close()
	}
}

func ShutdownCache() {
	engine := getEnv("CACHE_ENGINE", "local")

	if engine == "redis" {
		CloseRedis()
	} else {
		log.Println("🧹 Очистка локального кеша (папка cache)...")
		// Очищаємо всі файли в папці cache
		files, err := os.ReadDir("cache")
		if err != nil {
			return
		}
		for _, f := range files {
			os.Remove("cache/" + f.Name())
		}
	}
}
