package handlers

import (
	"florenbot/engine"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

// HandleClan обрабатывает команду /clan
func HandleClan(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: create, get, list, delete, update, join, leave"))
		return
	}

	action := parts[0]
	user_id := uint64(message.From.ID)

	switch action {
	case "create":
		{
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan create [имя]"))
				return
			}

			name := strings.Join(parts[1:], " ")
			engine.CreateClan(name, user_id)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Клан создан: %s", name)))
			return
		}
	case "delete":
		{
			clan_id, err := engine.GetClanByOwnerID(user_id)
			if err != nil {
				log.Printf("Ошибка получения клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан удален"))
				return
			}
	

			err = engine.DeleteClan(clan_id)
			if err != nil {
				log.Printf("Ошибка удаления клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Клан удален"))
			return
		}
	case "get":
    {
        // 1. Шукаємо, в якому клані користувач (як учасник)
        clan_id, err := engine.GetUserClanID(user_id)
        if err != nil {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
            return
        }

        // 2. Отримуємо дані клану
        clan, err := engine.GetClan(clan_id)
        if err != nil {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан не найден"))
            return
        }

        count, _ := engine.GetClanMemberCount(clan.Id)

        info := fmt.Sprintf("✅ *Информация о клане:*\n\n"+
            "🆔 ID: `%d`\n"+
            "🏷 Название: *%s*\n"+
            "👥 Участников: `%d`\n"+
            "👑 Владелец ID: `%d`",
            clan.Id, clan.Name, count, clan.OwnerId)

        msg := tgbotapi.NewMessage(message.Chat.ID, info)
        msg.ParseMode = "Markdown"
        bot.Send(msg)
        return
    }
	case "join":
		{
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan join [ID клана]"))
				return
			}

			clan_id_row := parts[1]

			clan_id, err := strconv.ParseUint(clan_id_row, 10, 32)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan join [ID клана]"))
				return
			}

			err = engine.JoinClan(clan_id, user_id)
			if err != nil {
				log.Printf("Ошибка при присоединении к клану: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы присоединились к клану"))
			return

		}
	case "leave":
        {
            // 1. Ищем клан, в котором состоит пользователь (как участник)
            clan_id, err := engine.GetUserClanID(user_id)
            if err != nil {
                // Если ошибки нет в базе, но клан не найден, значит пользователь не в клане
                bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
                return
            }

            log.Printf("Пользователь %d выходит из клана %d", user_id, clan_id)

            // 2. Выполняем выход из клана
            err = engine.LeaveClan(clan_id, user_id)
            if err != nil {
                log.Printf("Ошибка при выходе из клана: %v", err)
                bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что-то пошло не так при выходе из клана..."))
                return
            }
            
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы успешно вышли из клана"))
            return
        }
	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: create, get, list, delete, update"))
			return
		}
	}

}
