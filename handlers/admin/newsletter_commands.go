package admin

import (
	helpers "florenbot/engine/helpers"
	std_helpers "florenbot/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleNewsLetter(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	role, err := helpers.GetRole(user_id)
	if err != nil {
		log.Printf("Ошибка получения роли: %v", err)
		return
	}

	if role != "admin" && role != "owner" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет доступа к этой команде"))
		return
	}
	args := message.CommandArguments()
	parts := strings.Fields(args)

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/newsletter [private] [текст]`"))
		return
	}

	chat_type := parts[0]
	debug_type := std_helpers.GetEnvBool("DEBUG", false)
	text := strings.Join(parts[1:], " ")

	switch chat_type {
	case "private":
		{
			user_id := int64(message.From.ID)
			users, err := helpers.GetUsers()
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователей: %v", err)
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка получения пользователей")

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка отправки уведомления: %v", err)
				}

				return
			}

			// Получаем user IDS

			for _, user := range users {
				if debug_type {
					log.Printf("Отправка сообщения пользователям %d", user.Id)
				}

				msg := tgbotapi.NewMessage(int64(user.Id), text)

				if _, err := bot.Send(msg); err != nil {
					if debug_type {
						log.Printf("Ошибка отправки уведомления: %v", err)
					}
					return
				}

				bot.Send(
					tgbotapi.NewMessage(
						user_id,
						"✅ Сообщение успешно отправлено всем пользователям"),
				)
			}
		}
	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/newsletter [private] [текст]`"))
			return
		}
	}
}
