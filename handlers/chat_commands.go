package handlers

import (
	helpers "florenbot/engine/helpers"
	structs "florenbot/engine/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleChat(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)

	action := parts[0]

	switch action {
	case "create":
		{

			user_id := uint64(message.From.ID)
			newChat := structs.Chat{
				ID:     uint(message.Chat.ID),
				Name:   message.Chat.Title,
				UserID: int64(user_id),
			}

			err := helpers.CreateChat(newChat)
			if err != nil {
				log.Printf("Ошибка создания чата: %v", err)
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Чат создан"))
		}
	case "delete":
		{
			err := helpers.DeleteChat(message.Chat.ID)
			if err != nil {
				log.Printf("Ошибка удаления чата: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Чат удален"))
		}
	case "get":
		{
			chat, err := helpers.GetChatById(message.Chat.ID)
			if err != nil {
				log.Printf("Ошибка получения чата: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, chat.Name))
		}
	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/chat [create|delete|get]`"))
		}
	}
}
