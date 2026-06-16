package handlers

import (
	helpers "florenbot/engine/helpers"
	std_helpers "florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

func HandleThanks(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {

	user_reply := message.ReplyToMessage
	if user_reply == nil || message.ReplyToMessage.From == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, ответьте на сообщение пользователя.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	reply_user_id := uint64(user_reply.From.ID)
	debug_type := std_helpers.GetEnvBool("DEBUG", false)

	_, err := helpers.GetUserByID(reply_user_id)

	if err != nil {
		if debug_type {
			log.Printf("Ошибка получения пользователя: %v", err)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	current_reputate, err := helpers.GetReputation(reply_user_id)

	if err != nil {
		if debug_type {
			log.Printf("Ошибка получения репутации: %v", err)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка получения репутации")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	total_reputate := current_reputate + rand.Intn(100)
	log.Println("Репутация пользователя", reply_user_id, "увеличена на 200 : ", total_reputate)
	msgText := "И тебе!"

	if err := helpers.UpdatePositiveReputation_2(reply_user_id, total_reputate); err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обновления репутации в базе")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	msg.ReplyToMessageID = message.MessageID
	bot.Send(msg)
}

func HandleReputation(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Список доступных действий для автоматической генерации сообщений
	AvailableActions := []string{"полож", "отриц"}

	// 1. Попередні перевірки (валідація)
	args := message.CommandArguments()
	parts := strings.Fields(args)

	debug_type := GetEnvBool("DEBUG", false)

	if message.Chat.Type == "private" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, используйте бота в группе.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if len(parts) < 2 {
		helpText := fmt.Sprintf("❌ Неверный формат! Используйте: `/rep [%s] [количество]`", strings.Join(AvailableActions, "|"))
		msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if message.ReplyToMessage == nil || message.ReplyToMessage.From == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, ответьте на сообщение пользователя.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	reputate, err := strconv.Atoi(parts[1])
	if err != nil || reputate <= 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректное положительное число для изменения репутации.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if _, err := helpers.GetUserByID(uint64(message.From.ID)); err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы (зарегистрируйтесь в боте)")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	reply_user_id := uint64(message.ReplyToMessage.From.ID)
	user_id := uint64(message.From.ID)

	if reply_user_id == user_id {
		if debug_type {
			log.Printf("Пользователь %d пытается изменить репутацию самому себе", user_id)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не можете изменять репутацию самому себе!")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	action := parts[0]

	switch action {
	case "отриц":
		{
			_, err := helpers.GetUserByID(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			current_reputate, err := helpers.GetReputation(reply_user_id)

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

			if err := helpers.UpdateNetagiveReputation(reply_user_id, total_reputate); err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обновления репутации в базе"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

		}
	case "полож":
		{
			_, err := helpers.GetUserByID(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			current_reputate, err := helpers.GetReputation(reply_user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения репутации: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка получения репутации"))
				return
			}

			total_reputate := current_reputate + reputate
			msgText := fmt.Sprintf("✅ Репутация пользователя %s увеличена на %d", message.ReplyToMessage.From.FirstName, reputate)

			if err := helpers.UpdatePositiveReputation_2(reply_user_id, total_reputate); err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обновления репутации в базе"))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)

		}
	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Использование: `/rep [%s] [количество]`", strings.Join(AvailableActions, ","))))
		}
	}
}
