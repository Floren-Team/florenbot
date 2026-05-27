package handlers

import (
	"strings"
	"log"
	"fmt"
	"florenbot/engine"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleClan обрабатывает команду /clan
func HandleClan(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)

	if len(parts) < 1 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: create, get, list, delete, update"))
		return
	}

	action := parts[0]
	user_id := uint64(message.From.ID)


	switch action {
		case "create": {
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan create [имя]"))
				return
			}

			name := strings.Join(parts[1:], " ")
			engine.CreateClan(name, user_id)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Клан создан: %s", name)))
			return
		}
		case "delete": {
			clan_id, err := engine.GetClanByOwnerID(user_id)
			if err != nil {
				log.Printf("Ошибка получения клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
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
		case "get": {
			clan_id, err := engine.GetClanByOwnerID(user_id)
			if err != nil {
				log.Printf("Ошибка получения ID клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			
			clan, err := engine.GetClan(clan_id)
			if err != nil {
				log.Printf("Ошибка получения данных клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан не найден"))
				return
			}
			
			// Формуємо розгорнуту інформацію
			// Використовуємо \n для перенесення рядків
			info := fmt.Sprintf("✅ *Информация о клане:*\n\n"+
				"🆔 ID: `%d`\n"+
				"🏷 Название: *%s*\n"+
				"👑 Владелец ID: `%d`", 
				clan.Id, clan.Name, clan.OwnerId)
			
			msg := tgbotapi.NewMessage(message.Chat.ID, info)
			msg.ParseMode = "Markdown" // Додаємо парсинг Markdown для гарного вигляду
			bot.Send(msg)
			return
		}
		default: {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: create, get, list, delete, update"))
			return
		}
	}

}