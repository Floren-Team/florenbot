package handlers

import (
	"florenbot/engine"
	"florenbot/helpers"
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
	const available_actions = "create, list, delete, update, join, leave, createcode, getcode, invite, delcode, revoke"
	user_id := uint64(message.From.ID)

	if len(parts) < 1 {
		// bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: "+available_actions))
		// return

		clan_id, err := engine.GetUserClanID(user_id)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
			return
		}

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

	action := parts[0]

	switch action {
	case "create":
		{
			_, err := engine.GetUserClanID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			if err == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы состоите в клане"))
				return
			}

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
			_, err := engine.GetUserClanID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			clan_id, err := engine.GetClanByOwnerID(user_id)
			if err != nil {
				log.Printf("Ошибка получения клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не являетесь владельцем клана"))
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
	case "invite":
		{
			_, err := engine.GetUserClanID(user_id)

			if err == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы состоите в клане"))
				return
			}

			if err.Error() != "sql: no rows in result set" {
				log.Printf("Ошибка при проверке клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при проверке вашего статуса"))
				return
			}

			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan invite [код]"))
				return
			}

			code := parts[1]
			clan_id, err := engine.GetInviteCodeClan(code)
			if err != nil {
				log.Printf("Ошибка при приглашении в клан: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клана с таким кодом не существует"))
				return
			}

			err = engine.JoinClan(clan_id, user_id)
			if err != nil {
				log.Printf("Ошибка при приглашении в клан: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы присоеденились в клан!"))

		}
	case "revoke":
		{
			clan_id, err := engine.GetUserClanID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			err = engine.RevokeInviteCode(clan_id)
			if err != nil {
				log.Printf("Ошибка при удалении клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			code, err := engine.GetClanInviteCode(clan_id)
			if err != nil {
				log.Printf("Ошибка при удалении приглашения: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			msgText := fmt.Sprintf("✅ Сообщите его друзьям: /clan invite %s", code)

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, msgText))
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Старый код приглашения удален"))

		}
	case "delcode":
		{
			clan_id, err := engine.GetUserClanID(user_id)

			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			_, err = engine.GetClanInviteCode(clan_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет кода приглашения"))
				return
			}

			err = engine.DeleteInviteCode(clan_id)
			if err != nil {
				log.Printf("Ошибка при удалении приглашения: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Приглашение удалено"))
		}
	case "createcode":
		{
			clan_id, err := engine.GetUserClanID(user_id)

			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			_, err = engine.GetClanInviteCode(clan_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			if err == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас уже есть приглашение"))
				return
			}

			invite_code := helpers.GenerateCode()
			err = engine.CreateInviteCode(clan_id, invite_code)

			if err != nil {
				log.Printf("Ошибка при создании приглашения: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваш код приглашения: `%s`, скиньте его своим участникам: /clan invite %s", invite_code, invite_code)))
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

			_, err = engine.GetClan(clan_id)
			if err != nil {
				log.Printf("Ошибка получения клана: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан не найден"))
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
	case "getcode":
		{
			clan_id, err := engine.GetUserClanID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			invite_code, err := engine.GetClanInviteCode(clan_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет приглашения"))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваш код для приглашения: `%s`\nСкиньте его своим участникам: /clan invite %s", invite_code, invite_code)))
		}
	default:
		{
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: "+available_actions))
			return
		}
	}
}
