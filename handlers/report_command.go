package handlers

import (
	helpers "florenbot/engine/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleReport(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	debug_type := GetEnvBool("DEBUG", false)

	user_id := uint64(message.From.ID)

	args := message.CommandArguments()
	parts := strings.Fields(args)

	var AvailableActions = []string{
		"delete",
		"create",
	}

	msg_no_actions := fmt.Sprintf("❌ Неверный формат! Используйте: `/report [%s]`", strings.Join(AvailableActions, "|"))

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg_no_actions))
		return
	}

	action := parts[0]

	switch action {
	case "create":
		{
			if debug_type {
				log.Printf("Создание отчета %s\n", args)
			}

			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg_no_actions))
				return
			}

			exists_user, err := helpers.GetUserByID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			if exists_user.Id == 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			text := strings.Join(parts[1:], " ")

			if debug_type {
				log.Printf("Текст отчета: %s\n", text)
			}

			exists, err := helpers.HasReport(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка проверки наличия отчета: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить наличие отчета"))
				return
			}

			if exists {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У тебя уже есть отчет, подожди пж пока он будет обработан"))
				return
			}

			err = helpers.CreateReport(user_id, text)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка создания отчета: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось создать отчет"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ок, брат, я тебя понял))"))
			return

		}
	case "delete":
		{
			exists_user, err := helpers.GetUserByID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения пользователя: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			if exists_user.Id == 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден"))
				return
			}

			exists, err := helpers.HasReport(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка проверки наличия отчета: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить наличие отчета"))
				return
			}

			if !exists {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У тебя нет отчета"))
				return
			}

			err = helpers.DeleteReport(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка удаления отчета: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось удалить отчет"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ок, брат, я тебя понял))"))
			return
		}

	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg_no_actions))
			return
		}
	}
}
