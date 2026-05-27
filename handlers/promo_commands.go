package handlers

import (
	"log"
	"strconv"
	"strings"

	"florenbot/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandlePromo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /promo create [код] [сумма]\nВсе действия: active, create, delete"))
		return
	}

	action := parts[0]

	switch action {
		case "create":
			if len(parts) < 3 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /promo create [код] [сумма]"))
				return
			}

			user_id := int64(message.From.ID)
			_, err_2 := engine.GetUser(user_id)
			if err_2 != nil {
				log.Printf("Ошибка получения пользователя: %v", err_2)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			code := parts[1]
			amount, err := strconv.Atoi(parts[2])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Сумма должна быть числом"))
				return
			}

			// Проверяем существование кода
			_, err = engine.GetCode(code)

			// Если ошибка не связана с отсутствием записи (sql: no rows), значит это системная ошибка
			if err != nil && err.Error() != "sql: no rows in result set" {
				log.Printf("Ошибка БД при получении кода: %v", err) // Пишем в лог для разработчика
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
				log.Printf("Ошибка при создании кода: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при создании"))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод создан"))

		case "delete":
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /promo delete [код]"))
				return
			}

			// Перевіряємо, чи користувач авторизований (ваша перевірка)
			userID := int64(message.From.ID)
			_, err := engine.GetUser(userID)
			if err != nil {
				log.Printf("Ошибка получения пользователя: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			code := parts[1]

			// Перевіряємо існування промокоду
			_, err = engine.GetCode(code)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Промокод не найден"))
				} else {
					log.Printf("Ошибка БД при поиске кода: %v", err)
					bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при запросе к БД"))
				}
				return
			}

			// Викликаємо функцію, яка видаляє код і очищує його у користувачів
			err = engine.DeleteCode(code)
			if err != nil {
				log.Printf("Ошибка при удалении: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при удалении промокода"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод '"+code+"' удален и аннулирован у пользователей."))
			
		case "active":
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /promo active [код]"))
				return
			}
			code := parts[1]

			user_id := int64(message.From.ID)
			_, err_1 := engine.GetUser(user_id)
			if err_1 != nil {
				log.Printf("Ошибка получения пользователя: %v", err_1)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			existingPromo, err_2 := engine.GetUserCode(user_id)
			if err_2 != nil {
				log.Printf("Ошибка получения кода пользователя: %v", err_2)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка БД"))
				return
			}

			if existingPromo != "" {
           	 	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас уже есть активированный промокод: "+existingPromo))
            	return
        	}

			
			



			err := engine.ActivateCode(message.Chat.ID, code)
			if err != nil {
				log.Printf("Ошибка активации: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: "+err.Error()))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Промокод активирован"))

		default:
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестное действие\nИспользуйте: /promo [create|delete|active]"))
		}
	

	
}