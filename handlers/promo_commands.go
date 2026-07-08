package handlers

import (
	cache "florenbot/engine/cache"
	helpers "florenbot/engine/helpers"
	engine "florenbot/engine/mysql"
	std_helpers "florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

type Promocode struct {
	Code   string
	Amount int
}

func HandlePromo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)
	user_id := uint64(message.From.ID)
	debug_type := std_helpers.GetEnvBool("DEBUG", false)
	userProfile, err := helpers.GetUserByID(user_id)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
		return
	}

	if userProfile.ID == 0 {
		if debug_type {
			log.Printf("Пользователь %s создается...", message.From.FirstName)
		}
		userProfile, err = helpers.CreateUser(
			user_id,
			message.From.UserName,
			message.From.FirstName,
		)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при создании пользователя: %v", err)
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось создать пользователя"))
			return
		}

		if debug_type {
			log.Printf("Пользователь %s создан", message.From.FirstName)
			log.Printf("Результат: %v", userProfile)
		}
	} else {
		if debug_type {
			log.Printf("Пользователь %s уже существует", message.From.FirstName)
		}
	}

	var availableActions = []string{"active", "create", "delete", "list", "stats"}

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /promo create [код] [сумма]\nАктивация промокода: /promo active [код]\nВсе действия: "+strings.Join(availableActions[:], ", ")))
		return
	}

	debug_type = std_helpers.GetEnvBool("DEBUG", false)

	action := parts[0]

	switch action {
	case "create":

		user_id := uint64(message.From.ID)
		balance, err := cache.GetBalance(user_id, message.From.UserName)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения баланса: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
			return
		}

		log.Printf("Баланс пользователя: %.2f", balance)

		if balance < 2000 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Недостаточно средств. Необходимо 2000 $"))
			return
		}

		if len(parts) < 3 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /promo create [код] [сумма]"))
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
		_, err = helpers.GetCode(code)

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
		err = helpers.CreateCode(code, amount, user_id)
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
		_, err := helpers.GetUserByID(userID)
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

		_, err = helpers.GetCode(code)
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

		err = helpers.DeleteCode(code)
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
		_, err := helpers.GetUserByID(user_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения пользователя: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
			return
		}

		// 2. Проверка, не активирован ли уже код
		existingPromo, err := helpers.GetUserCode(user_id)
		if err != nil || existingPromo != "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас уже есть активированный промокод или ошибка БД"))
			return
		}

		// 3. Получение суммы промокода
		promo, err := helpers.GetCode(code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка БД при поиске кода: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден"))
			return
		}

		amount := promo.Amount

		// 4. Активация: записываем код пользователю (только в БД)
		err = helpers.ActivateCode(user_id, code)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при активации: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при активации"))
			return
		}

		// 5. ОБНОВЛЕНИЕ БАЛАНСА С КЭШЕМ
		// Сначала получаем текущий баланс (из кэша, если есть, или из БД)
		currentBalance, err := cache.GetBalance(user_id, message.From.UserName)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения баланса: %v", err)
			}
		}

		// Рассчитываем новый баланс
		newBalance := currentBalance + amount

		// Обновляем БД
		err = engine.UpdateBalance(user_id, amount)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка обновления БД: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при начислении бонуса"))
			return
		}

		// Обновляем КЭШ напрямую
		key := fmt.Sprintf("user_%d_balance", user_id)
		cache.SetCache(key, strconv.FormatFloat(newBalance, 'f', 2, 64), 24*time.Hour)

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Промокод активирован! Вам начислено +%.2f $. Ваш баланс: %2.f", amount, newBalance)))

	case "list":
		user_id := uint64(message.From.ID)
		promocodes, err := helpers.GetPromocodesUser(user_id)

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
			sb.WriteString(fmt.Sprintf("├── `%s` — %2.f $\n", p.Code, p.Amount))
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "edit":
		{
			// 1. Проверка пользователя
			user_id := uint64(message.From.ID)
			_, err := helpers.GetUserByID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			args = message.CommandArguments()
			parts = strings.Fields(args)

			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/promo edit [код] [сумма]`"))
				return
			}

			code := parts[0]
			amount, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/promo edit [код] [сумма]`"))
				return
			}

			code_db, err := helpers.GetUserCode(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении кода: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении кода"))
				return
			}

			if code_db != code {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный код"))
				return
			}

			// Обновляем БД
			err = helpers.UpdatePromo(code, amount, user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка обновления БД: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при обновлении промокода"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Промокод обновлен!\nСумма: %.2f\nНовое имя: %s", amount, code))
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	case "stats":
		{
			count, err := helpers.GetPromocodesMemberCount()
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении количества участников промокодов: %v", err)
				}
				log.Printf("Ошибка при получении количества участников промокодов: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении количества участников промокодов"))
				return
			}

			if len(count) == 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет активных промокодов"))
				return
			}

			var report strings.Builder
			report.WriteString("📊 **Статистика промокодов**\n\n")
			report.WriteString("```\n")
			report.WriteString(fmt.Sprintf("%-12s | %-10s\n", "Промокод", "Участники"))
			report.WriteString("---------------------------\n")
			for code, num := range count {
				report.WriteString(fmt.Sprintf("%-12s | %-10d\n", code, num))
			}
			report.WriteString("```")
			msg := tgbotapi.NewMessage(message.Chat.ID, report.String())
			msg.ParseMode = "Markdown"
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
			}

		}
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестное действие\nВсе действия: "+strings.Join(availableActions[:], ", ")))
	}

}
