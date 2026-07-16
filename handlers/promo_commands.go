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

	"context"
	"encoding/json"
	"errors"
	structs "florenbot/engine/structs"
	"gorm.io/gorm"
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

	var availableActions = []string{"active", "create", "delete", "list", "stats", "delexpire", "expire"}

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
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Сумма должна быть числом"))
			return
		}

		_, err = helpers.GetCode(code)

		if err == nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод уже существует"))
			return
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			if debug_type {
				log.Printf("Ошибка БД при проверке кода: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при запросе к БД"))
			return
		}

		err = helpers.CreateCode(code, amount, user_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при создании кода: %v", err)
			}
			errorMessage := fmt.Sprintf("❌ Ошибка при создании кода: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, errorMessage))
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод создан")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)

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

		promo, err := helpers.GetCode(code)
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

		result, err := std_helpers.IsCreator(bot, message.Chat.ID, int64(message.From.ID))
		log.Printf("result checker creator: %v", result)
		log.Printf("error checker creator: %v", err)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при проверке создателя: %v", err)
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при проверке создателя")
			msg.ReplyToMessageID = message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
			}
			return
		}

		if result {
			if debug_type {
				log.Printf("Удаление промокода %s", code)
			}

			err = helpers.DeleteCode(code)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при удалении промокода"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод '"+code+"' удален\n"+
														 "❌ Промокод аннулирован у всех пользователей.\n"+
														 "Кто удалил: Создатель"))
			return
		}
		if uint64(promo.OwnerID) != userID {
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не можете удалить этот промокод")
			msg.ReplyToMessageID = message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки уведомления: %v", err)
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
			user_id := uint64(message.From.ID)
			_, err := helpers.GetUserByID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			args := message.CommandArguments()
			parts := strings.Fields(args)

			// Учитываем, что parts[0] - подкоманда "edit", если CommandArguments возвращает [edit code amount]
			if len(parts) < 3 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/promo edit [код] [сумма]`"))
				return
			}

			code := parts[1]
			amount, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: сумма должна быть числом"))
				return
			}

			// Проверка владельца: получаем структуру кода из БД
			promo, err := helpers.GetCode(code) // Предполагается, что функция возвращает структуру с OwnerID
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден"))
				return
			}

			// СРАВНИВАЕМ ВЛАДЕЛЬЦА
			if uint64(promo.OwnerID) != user_id {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец этого промокода"))
				return
			}

			// Обновляем БД
			err = helpers.UpdatePromo(code, amount, user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при обновлении"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Промокод *%s* обновлен!\nНовая сумма: *%.2f*", code, amount))
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	case "expire":
		{
			// 1. Проверка пользователя
			user_id := uint64(message.From.ID)
			_, err := helpers.GetUserByID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			// Получаем аргументы команды
			args := message.CommandArguments()
			parts := strings.Fields(args)

			if len(parts) < 3 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат!\nИспользуйте: `/promo expire [код] [время]`"))
				return
			}

			code := parts[1]
			timeStr := parts[2]

			promo, err := helpers.GetCode(code)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный код"))
				return
			}

			if uint64(promo.OwnerID) != user_id {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец этого кода"))
				return
			}

			if !promo.ExpiresAt.IsZero() {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У этого промокода уже установлен срок действия"))
				return
			}

			// Парсим длительность
			duration, err := std_helpers.ParseDuration(timeStr)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка формата времени: "+timeStr+"\nИспользуйте формат (d, h, m)\nПример: `1h`, `30d`"))
				return
			}

			// Вычисляем время для БД
			expireTime := time.Now().Add(duration)
			expiresAtStr := expireTime.Format("2006-01-02 15:04:05")

			// 1. Обновляем БД
			err = helpers.UpdatePromoExpire(code, expiresAtStr)
			if err != nil {
				log.Printf("Ошибка обновления БД для кода %s: %v", code, err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при обновлении в базе данных"))
				return
			}

			timerData := structs.PromoTimer{
				UserID:      int64(user_id),
				PromoString: code,
			}
			data, _ := json.Marshal(timerData)

			redisKey := "promo_expire:" + code
			err = cache.GetRedis().Set(context.Background(), redisKey, data, duration).Err()
			if err != nil {
				log.Printf("Ошибка записи в Redis: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка системы таймера"))
				return
			}

			// Успех
			formattedDatetime := expireTime.Format("02.01.2006 15:04")
			msgText := fmt.Sprintf("✅ Промокод успешно обновлен!\nКод: *%s*\nСрок действия до: *%s*", code, formattedDatetime)

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

	case "delexpire":
		{
			user_id := uint64(message.From.ID)

			// 1. Авторизация
			_, err := helpers.GetUserByID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			// 2. Получение аргументов команды
			args := message.CommandArguments()
			parts := strings.Fields(args)

			if len(parts) < 1 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат!\nИспользуйте: `/promo delexpire [код]`"))
				return
			}

			code := parts[0]

			// 3. Получение данных промокода из БД
			promo, err := helpers.GetCode(code)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден"))
				return
			}

			// 4. Проверка прав владельца
			if uint64(promo.OwnerID) != user_id {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец этого кода"))
				return
			}

			// 5. Проверка, есть ли срок действия
			if promo.ExpiresAt.IsZero() {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Этот промокод не имеет срока действия"))
				return
			}

			// 6. Удаление из Redis с использованием составного ключа
			redisKey := fmt.Sprintf("promo_expire:%s:%d", code, user_id)

			rdb := cache.GetRedis()
			if rdb == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка системы: Redis недоступен"))
				return
			}

			// Удаляем ключ и проверяем результат
			deletedCount, err := rdb.Del(context.Background(), redisKey).Result()
			if err != nil {
				log.Printf("Ошибка удаления из Redis для кода %s: %v", code, err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при удалении таймера в системе"))
				return
			}

			// 7. Удаление срока действия в БД
			err = helpers.DeletePromoExpire(code)
			if err != nil {
				log.Printf("Ошибка удаления срока для кода %s: %v", code, err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при удалении срока в БД"))
				return
			}

			// 8. Формирование ответа
			responseText := fmt.Sprintf("✅ Срок действия промокода *%s* успешно удален!", code)
			if deletedCount == 0 {
				responseText += "\n_(Таймер уже был завершен или не существовал)_"
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

	case "stats":
		count, err := helpers.GetPromocodesMemberCount()
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при получении количества участников промокодов: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении статистики"))
			return
		}

		if len(count) == 0 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Нет активных промокодов"))
			return
		}

		var report strings.Builder
		report.WriteString("📊 **Статистика промокодов**\n\n")
		report.WriteString("```\n")
		report.WriteString(fmt.Sprintf("%-12s | %-10s\n", "Промокод", "Участники"))
		report.WriteString("---------------------------\n")

		for code, num := range count {
			displayCode := code
			if displayCode == "" {
				displayCode = "N/A"
			}
			report.WriteString(fmt.Sprintf("%-12s | %-10d\n", displayCode, num))
		}
		report.WriteString("```")

		msg := tgbotapi.NewMessage(message.Chat.ID, report.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестное действие\nВсе действия: "+strings.Join(availableActions[:], ", ")))
	}

}
