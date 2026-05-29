package handlers

import (
	"florenbot/engine"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func getEnvBool(key string, defaultValue bool) bool {
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

func HandleReputation(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Список доступных действий для автоматической генерации сообщений
	AvailableActions := []string{"полож", "отриц"}
	
	// 1. Попередні перевірки (валідація)
	args := message.CommandArguments()
	parts := strings.Fields(args)

	debug_type := GetEnvBool("DEBUG", false)

	if message.Chat.Type == "private" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, используйте бота в группе."))
		return
	}

	if len(parts) < 2 {
		helpText := fmt.Sprintf("❌ Неверный формат! Используйте: `/rep [%s] [количество]`", strings.Join(AvailableActions, "|"))
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, helpText))
		return
	}



	if message.ReplyToMessage == nil || message.ReplyToMessage.From == nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Пожалуйста, ответьте на сообщение пользователя."))
		return
	}

	reputate, err := strconv.Atoi(parts[1])
	if err != nil || reputate <= 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректное положительное число для изменения репутации."))
		return
	}

	if _, err := engine.GetUserByID(uint64(message.From.ID)); err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы (зарегистрируйтесь в боте)"))
		return
	}

	reply_user_id := uint64(message.ReplyToMessage.From.ID)
	user_id := uint64(message.From.ID)

	if reply_user_id == user_id {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не можете изменять репутацию самому себе!"))
		return
	}

	action := parts[0]

	switch action {
		case "отриц": {
			_, err := engine.GetUserByID(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			current_reputate, err := engine.GetReputation(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения репутации: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка получения репутации"))
				return
			}

			if current_reputate <= 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У пользователя нет репутации для уменьшения"))
				return
			}

			total_reputate := current_reputate - reputate
			msgText := fmt.Sprintf("✅ Репутация пользователя %s уменьшена на %d", message.ReplyToMessage.From.FirstName, reputate)
	

			if err := engine.UpdateNetagiveReputation(reply_user_id, total_reputate); err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обновления репутации в базе"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

		}
		case "полож": {
			_, err := engine.GetUserByID(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			current_reputate, err := engine.GetReputation(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения репутации: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка получения репутации"))
				return
			}

			

			total_reputate := current_reputate + reputate
			msgText := fmt.Sprintf("✅ Репутация пользователя %s увеличена на %d", message.ReplyToMessage.From.FirstName, reputate)
	

			if err := engine.UpdatePositiveReputation_2(reply_user_id, total_reputate); err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обновления репутации в базе"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

		}
		default: {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Использование: `/rep [%s] [количество]`", strings.Join(AvailableActions, ","))))	
		}
	}
}