package handlers

import (
	"florenbot/engine"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Promocode struct {
	Code   string
	Amount int
}

func GetEnvBool_2(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func HandlePromo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /promo create [код] [сумма]\nВсе действия: active, create, delete, list"))
		return
	}

	debug_type := GetEnvBool_2("DEBUG", false)

	action := parts[0]

	switch action {
	case "create":

		user_id := uint64(message.From.ID)
		balance, err := engine.GetBalance(user_id, message.From.UserName)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения баланса: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
			return
		}

		if balance < 2000 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Недостаточно средств. Необходимо 2000 рублей"))
			return
		}

		if len(parts) < 3 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /promo create [код] [сумма]"))
			return
		}

		_, err_2 := engine.GetUserByID(user_id)
		if err_2 != nil {
			if debug_type {
				log.Printf("Ошибка получения пользователя: %v", err_2)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
			return
		}

		code := parts[1]
		amount, err := strconv.Atoi(parts[2])
		if err != nil {
			if debug_type {
				log.Printf("Ошибка парсинга суммы: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Сумма должна быть числом"))
			return
		}

		// Проверяем существование кода
		_, err = engine.GetCode(code)

		// Если ошибка не связана с отсутствием записи (sql: no rows), значит это системная ошибка
		if err != nil && err.Error() != "sql: no rows in result set" {
			if debug_type {
				log.Printf("Ошибка БД при получении кода: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при запросе к БД"))
			return
		}

		// Если err == nil, значит код уже есть
		if err == nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод уже существует"))
			return
		}

		// Создаем код
		err = engine.CreateCode(code, amount)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при создании кода: %v", err)
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при создании")
			msg.ReplyToMessageID = message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
			}
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод создан")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}

	case "delete":
		if len(parts) < 2 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /promo delete [код]")
			msg.ReplyToMessageID = message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
			}
			return
		}

		userID := uint64(message.From.ID)
		_, err := engine.GetUserByID(userID)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения пользователя: %v", err)
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы")
			msg.ReplyToMessageID = message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
			}
			return
		}

		code := parts[1]

		_, err = engine.GetCode(code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка БД при поиске кода: %v", err)
			}
			if err.Error() == "sql: no rows in result set" {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден")
				msg.ReplyToMessageID = message.MessageID
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка отправки уведомления: %v", err)
				}
			} else {
				log.Printf("Ошибка БД при поиске кода: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при запросе к БД"))
			}
			return
		}

		err = engine.DeleteCode(code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при удалении: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при удалении промокода"))
			return
		}

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод '"+code+"' удален и аннулирован у пользователей."))

	case "active":
		if len(parts) < 2 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите код\nИспользуйте: /promo active [код]"))
			return
		}
		code := parts[1]
		user_id := uint64(message.From.ID)

		// 1. Проверка авторизации
		_, err := engine.GetUserByID(user_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения пользователя: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
			return
		}

		// 2. Проверка, не активирован ли уже код
		existingPromo, err := engine.GetUserCode(user_id)
		if err != nil || existingPromo != "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас уже есть активированный промокод или ошибка БД"))
			return
		}

		// 3. Получение суммы промокода
		amount, err := engine.GetCode(code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка БД при поиске кода: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден"))
			return
		}

		// 4. Активация: записываем код пользователю (только в БД)
		err = engine.ActivateCode(user_id, code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при активации: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при активации"))
			return
		}

		// 5. ОБНОВЛЕНИЕ БАЛАНСА С КЭШЕМ
		// Сначала получаем текущий баланс (из кэша, если есть, или из БД)
		currentBalance, err := engine.GetBalance(user_id, message.From.UserName)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения баланса: %v", err)
			}
		}

		// Рассчитываем новый баланс
		newBalance := currentBalance + amount

		// Обновляем БД
		err = engine.UpdateBalanceSQL(user_id, amount)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка обновления БД: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при начислении бонуса"))
			return
		}

		// Обновляем КЭШ напрямую
		key := fmt.Sprintf("user_%d_balance", user_id)
		engine.SetCache(key, strconv.Itoa(newBalance), 24*time.Hour)

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Промокод активирован! Вам начислено +%d рублей. Ваш баланс: %d", amount, newBalance)))

	case "list":
		user_id := uint64(message.From.ID)
		promocodes, err := engine.GetPromocodesUser(user_id)

		if err != nil {
			if debug_type {
				log.Printf("Ошибка при получении промокодов: %v", err)
			}
			log.Printf("Ошибка при получении промокодов: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении промокодов"))
			return
		}

		if len(promocodes) == 0 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет активных промокодов"))
			return
		}

		var sb strings.Builder
		sb.WriteString("📋 **Ваши промокоды:**\n\n")

		for _, p := range promocodes {
			sb.WriteString(fmt.Sprintf("├── `%s` — %d монет\n", p.Code, p.Amount))
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестное действие\nИспользуйте: /promo [create|delete|active|list]"))
	}

}
