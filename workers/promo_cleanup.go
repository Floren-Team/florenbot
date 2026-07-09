package workers

import (
    "context"
    "log"
    "strings"

    cache "florenbot/engine/cache"
    "florenbot/engine/helpers"
    "time"
    "github.com/redis/go-redis/v9"
)

// RunPromoCleanupWorker слушает события истечения срока действия ключей в Redis
func RunPromoCleanupWorker() {
    for {
        client := cache.GetRedis()
        if client == nil {
            log.Println("⚠️ Redis не инициализирован, жду 5 секунд...")
            time.Sleep(5 * time.Second)
            continue
        }

        ctx := context.Background()
        pubsub := client.PSubscribe(ctx, "__keyevent@0__:expired")
        
        log.Println("✅ Воркер очистки запущен")

        // Используем Receive, чтобы получить сообщение
        for {
            msg, err := pubsub.Receive(ctx)
            if err != nil {
                // Если ошибка в том, что клиент закрыт — выходим из цикла
                if err.Error() == "redis: client is closed" {
                    log.Println("🛑 Redis клиент закрыт, останавливаю воркер")
                    return 
                }
                log.Printf("❌ Ошибка воркера: %v", err)
                // Небольшая задержка перед переподключением
                time.Sleep(2 * time.Second)
                break 
            }

            // Проверяем тип сообщения
            switch m := msg.(type) {
            case *redis.Message:
                // Проверяем, что это нужный нам ключ
                if strings.HasPrefix(m.Payload, "promo_expire:") {
                    code := strings.TrimPrefix(m.Payload, "promo_expire:")
                    
                    log.Printf("⏳ Время истекло для кода: %s. Удаляю из БД...", code)
                    
                    err := helpers.DeleteCode(code)
                    if err != nil {
                        log.Printf("❌ Ошибка при удалении %s из БД: %v", code, err)
                    } else {
                        log.Printf("✅ Промокод %s успешно удален из БД", code)
                    }
                }
            case *redis.Subscription:
                // Игнорируем сообщения о подтверждении подписки
                continue
            }
        }
        pubsub.Close()
    }
}