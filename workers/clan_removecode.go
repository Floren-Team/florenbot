package workers

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"florenbot/engine/cache"
	"florenbot/engine/helpers"

	"github.com/redis/go-redis/v9"
)

// ClanRemoveInviteCodeWorker слушает события истечения срока действия ключей в Redis
func ClanRemoveInviteCodeWorker() {
	for {
		client := cache.GetRedis()
		if client == nil {
			log.Println("⚠️ Redis не инициализирован, жду 5 секунд...")
			time.Sleep(5 * time.Second)
			continue
		}

		ctx := context.Background()
		pubsub := client.PSubscribe(ctx, "__keyevent@0__:expired")

		log.Println("✅ Воркер очистки кода приглашеня кланов запущен")

		for {
			msg, err := pubsub.Receive(ctx)
			if err != nil {
				if err.Error() == "redis: client is closed" {
					log.Println("🛑 Redis клиент закрыт, останавливаю воркер")
					return
				}
				log.Printf("❌ Ошибка воркера: %v", err)
				time.Sleep(2 * time.Second)
				break
			}

			switch m := msg.(type) {
			case *redis.Message:
				// 1. Обработка промокодов
				if strings.HasPrefix(m.Payload, "promo_expire:") {
					code := strings.TrimPrefix(m.Payload, "promo_expire:")
					log.Printf("⏳ Время истекло для промокода: %s. Удаляю из БД...", code)

					err := helpers.DeleteCode(code)
					if err != nil {
						log.Printf("❌ Ошибка при удалении %s из БД: %v", code, err)
					} else {
						log.Printf("✅ Промокод %s успешно удален", code)
					}
				}

				// 2. Обработка приглашений в клан
				if strings.HasPrefix(m.Payload, "clan_invite:") {
					clanIDStr := strings.TrimPrefix(m.Payload, "clan_invite:")
					clanID, err := strconv.ParseInt(clanIDStr, 10, 64)
					if err != nil {
						log.Printf("❌ Ошибка парсинга ID клана из Redis: %v", err)
						continue
					}

					log.Printf("⏳ Время истекло для клана ID: %d. Удаляю приглашение...", clanID)
					err = helpers.DeleteInviteCode(uint64(clanID))
					if err != nil {
						log.Printf("❌ Ошибка при удалении приглашения клана %d из БД: %v", clanID, err)
					} else {
						log.Printf("✅ Приглашение клана %d успешно удалено", clanID)
					}
				}

			case *redis.Subscription:
				continue
			}
		}
		pubsub.Close()
	}
}