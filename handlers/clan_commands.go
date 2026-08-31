package handlers

import (
	"context"
	"encoding/json"
	"errors"
	cache "florenbot/engine/cache"
	helpers "florenbot/engine/helpers"
	structs "florenbot/engine/structs"
	helper "florenbot/helpers"
	std_helpers "florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

// HandleClan обрабатывает команду /clan
func HandleClan(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)
	var AvailableActions = []string{
		"create", "list", "delete", "update", "join",
		"leave", "createcode", "getcode", "invite",
		"delcode", "revoke", "kick", "add", "ban", "expirecode", "delexpire",
	}
	user_id := uint64(message.From.ID)
	debug_type := GetEnvBool("DEBUG", false)

	if len(parts) < 1 {
		clan_id, err := helpers.GetUserClanID(user_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения клана: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
			return
		}

		clan, err := helpers.GetClanByID(clan_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения клана: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан не найден"))
			return
		}

		count, _ := helpers.GetClanMemberCount(clan_id)

		info := fmt.Sprintf("✅ *Информация о клане:*\n\n"+
			"🆔 ID: `%d`\n"+
			"🏷 Название: *%s*\n"+
			"👥 Участников: `%d`\n"+
			"👑 Имя владельца: `%s`",
			clan.ID, clan.Name, count, clan.OwnerName)

		msg := tgbotapi.NewMessage(message.Chat.ID, info)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	action := parts[0]

	switch action {
	case "create":
		{
			// Пытаемся получить ID клана пользователя
			_, err := helpers.GetUserClanID(user_id)

			// 1. Проверяем: произошла ли ошибка
			if err != nil {
				// Если это НЕ "запись не найдена" — значит, это реальная ошибка БД
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					if debug_type {
						log.Printf("Ошибка получения клана из БД: %v", err)
					}
					bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при обращении к базе данных"))
					return
				}
				// Если err == gorm.ErrRecordNotFound, мы просто продолжаем выполнение — пользователь без клана
			} else {
				// Если err == nil, значит запись найдена (клан есть)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы уже состоите в клане"))
				return
			}

			// ... далее ваш код создания клана (баланс, проверка формата и т.д.)
			user_id := uint64(message.From.ID)
			balance, err := cache.GetBalance(user_id, message.From.UserName)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			if balance < 1000 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Недостаточно средств. Необходимо 1000 рублей"))
				return
			}

			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan create [имя]"))
				return
			}

			// ... создание клана
			name := strings.Join(parts[1:], " ")
			err = helpers.CreateClan(name, message.From.FirstName, user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось создать клан"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Клан создан: %s", name)))
			return
		}
	case "getrole":
		{
			_, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваша роль в клане: %s", role)))
			return
		}
	case "kick":
		{

			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			if role != "admin" && role != "owner" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет прав администратора/владельца клана"))
				return
			}

			// Получить reply message пользователя

			reply := message.ReplyToMessage
			if reply == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, ответьте на сообщение пользователя который собираетесь кикнуть"))
				return
			}

			user_reply_id_raw := uint64(reply.From.ID)

			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan kick [причина]"))
				return
			}

			reason := strings.Join(parts[1:], " ")

			if len(reason) > 100 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Причина не должна быть длиннее 100 символов"))
				return
			} else if len(reason) < 3 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Причина не должна быть короче 3 символов"))
				return
			}

			user_reply_id := int64(user_reply_id_raw)

			_, err = helpers.GetUserClanID(user_reply_id_raw)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не состоит в клане")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			// Кикнуть пользователя
			if err := helpers.KickClanUser(clan_id, user_reply_id_raw); err != nil {
				if debug_type {
					log.Printf("Ошибка кикануть пользователя: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось кикнуть пользователя"))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Пользователь %s кикнут из клана", reply.From.FirstName)))

			// Ответить пользователю
			msg := tgbotapi.NewMessage(user_reply_id, fmt.Sprintf("✅ Вы были кикнуты из клана: %s", reason))
			bot.Send(msg)
			return
		}
	case "list":
		{
			clans, err := helpers.GetClans()
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения кланов: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			if len(clans) == 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Кланов нет в этом боте!"))
				return
			}

			info := "✅ *Список кланов:*\n\n"
			for _, clan := range clans {
				clanDetails := fmt.Sprintf(
					"🆔 ID: `%d`\n🏷 Название: *%s*\n👥 Участников: `%d`",
					clan.ID, clan.Name, clan.MemberCount,
				)

				info += clanDetails
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, info)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
			return
		}
	case "ban":
		{

			clan_id, err := helpers.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения роли: %v", err)
				}
				log.Printf("Ошибка получения роли: %v", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так...")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}
			if debug_type {
				log.Printf("Роль: %s", role)
			}
			if role != "admin" && role != "owner" {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет прав администратора/владельца клана")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			if len(parts) < 2 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan ban [причина]")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			user_reply := message.ReplyToMessage
			if user_reply == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пожалуйста, ответьте на сообщение пользователя которого хотите забанить")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			user_reply_id := uint64(user_reply.From.ID)

			_, err = helpers.GetUserClanID(user_reply_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не состоит в клане")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			reason := strings.Join(parts[1:], " ")

			if len(reason) > 100 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Причина не должна быть длиннее 100 символов")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			} else if len(reason) < 3 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Причина не должна быть короче 3 символов")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}

			// Забанить пользователя
			err = helpers.BlockMemberClan(clan_id, user_reply_id, reason)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка бана ЮЗЕРА: %v", err)
				}
				log.Printf("Ошибка блокировки пользователя: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Пользователь %s забанен", user_reply.From.FirstName))
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}
	case "delete":
		{
			_, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			clan_id, err := helpers.GetClanByOwnerID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не являетесь владельцем клана"))
				return
			}

			err = helpers.DeleteClan(uint64(clan_id.ID))
			if err != nil {
				if debug_type {
					log.Printf("Ошибка удаления клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Клан удален"))
			return
		}
	case "invite":
		{
			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan invite недоступна в группах"))
				return
			}

			_, err := helpers.GetUserClanID(user_id)

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
			clan, err := helpers.GetClanByInviteCode(code)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клана с таким кодом не существует"))
				return
			}
			clan_id := uint64(clan.ID)
			err = helpers.CheckBlacklist(clan_id, user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вас заблокировали!"))
				return
			}

			err = helpers.JoinClan(clan_id, user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы присоеденились в клан!"))

			// Отправляем владельцу сообщение о приглашении
			owner_id_raw, err := helpers.GetClanOwnerID(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			owner_id := int64(owner_id_raw)

			bot.Send(tgbotapi.NewMessage(owner_id, fmt.Sprintf("К вам зашел пользователь %s", message.From.FirstName)))

		}
	case "revoke":
		{

			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan revoke недоступна в группах"))
				return
			}

			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			if role != "owner" && role != "admin" {
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			err = helpers.RevokeInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			newCode, err := helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении нового кода: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			// Відправляємо повідомлення
			_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Старый код приглашения удален"))
			_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Новый код: /clan invite %s", newCode)))

		}
	case "delcode":
		{
			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan delcode недоступна в группах"))
			}
			clan_id, err := helpers.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			if role != "owner" && role != "admin" {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			code, err := helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет кода приглашения"))
				return
			}

			if code == "" {
				if debug_type {
					log.Printf("Ошибка при удалении приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет кода приглашения"))
				return
			}

			err = helpers.DeleteInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Приглашение удалено"))
		}
	case "add":
		{
			if message.Chat.Type == "private" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan add недоступна в ЛС!"))
				return
			}

			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			reply_msg := message.ReplyToMessage
			if reply_msg == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ответьте на сообщение пользователя"))
				return
			}

			target_user_id := uint64(reply_msg.From.ID)

			existing_clan, _ := helpers.GetUserClanID(target_user_id)
			if existing_clan != 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь уже состоит в клане"))
				return
			}

			err = helpers.AddUserToClan(clan_id, target_user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Пользователь добавлен в клан"))
			reply_user_id := int64(reply_msg.From.ID)
			bot.Send(tgbotapi.NewMessage(reply_user_id, fmt.Sprintf("✅ Вас добавил %s", message.From.FirstName)))
		}
	case "createcode":
		{
			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan createcode недоступна в группах"))
			}
			clan_id, err := helpers.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			if role != "owner" && role != "admin" {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			_, err = helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			invite_code := helper.GenerateCode()
			err = helpers.CreateInviteCode(clan_id, invite_code)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваш код приглашения: `%s`, скиньте его своим участникам: /clan invite %s", invite_code, invite_code)))
		}
	case "expirecode":
		{
			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan expirecode недоступна в группах"))
			}

			args := message.CommandArguments()
			parts := strings.Fields(args)

			// 1. Проверяем, что есть хотя бы подкоманда (индекс 0) и аргумент времени (индекс 1)
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: `/clan expirecode [время]`"))
				return
			}

			// Теперь безопасно обращаться к индексу 1
			rawDuration := parts[1]
			log.Println("Получил действие: ", parts[0], " время: ", rawDuration)

			duration, err := std_helpers.ParseDuration(rawDuration)
			if err != nil {
				if debug_type {
					log.Printf("Error: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверное время! Используйте: `/clan expirecode [время]`"))
				return
			}

			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			if role != "owner" && role != "admin" {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			code, err := helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			if code == "" {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У клана нет приглашения"))
				return
			}

			_, err = helpers.GetClanInviteCode(clan_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			timerData := structs.ClanInviteTimer{
				ClanID: int64(clan_id),
				UserID: int64(user_id),
			}
			data, _ := json.Marshal(timerData)

			redisKey := "clan_invite:" + strconv.FormatInt(int64(clan_id), 10)
			err = cache.GetRedis().Set(context.Background(), redisKey, data, duration).Err()
			if err != nil {
				log.Printf("Ошибка записи в Redis: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка системы таймера"))
				return
			}

			expireTime := time.Now().Add(duration)
			expiresAtStr := expireTime.Format("2006-01-02 15:04:05")

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Приглашение будет удалено через %s\nОтменить: /clan delexpire", expiresAtStr)))
		}
	case "delexpire":
		{
			// 1. Проверка на группы с обязательным return
			if message.Chat.Type == "supergroup" || message.Chat.Type == "group" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Команда /clan delexpire недоступна в группах"))
				return // Обязательно выходим, чтобы код не шел дальше
			}

			// 2. Получение ID клана
			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			// 3. Проверка роли
			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при проверке роли: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка проверки прав"))
				return
			}

			if role != "owner" && role != "admin" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			// 4. Проверка существования приглашения в БД
			code, err := helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении кода приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет активного приглашения"))
				return
			}

			if code == "" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У клана нет приглашения"))
				return
			}

			// 5. Удаление таймера из Redis
			redisKey := "clan_invite:" + strconv.FormatInt(int64(clan_id), 10)

			// Получаем клиент Redis через функцию из пакета cache
			rdb := cache.GetRedis()
			if rdb == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: Redis недоступен"))
				return
			}

			deletedCount, err := rdb.Del(context.Background(), redisKey).Result()
			if err != nil {
				log.Printf("Ошибка при работе с Redis: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка системы таймера"))
				return
			}

			// 6. Обработка результата удаления
			if deletedCount == 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Нет активного таймера. Пожалуйста, инициализируйте его через: /clan expirecode [время]"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Срок действия приглашения успешно удален"))
		}
	case "leave":
		{
			// 1. Ищем клан, в котором состоит пользователь (как участник)
			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении клана: %v", err)
				}
				// Если ошибки нет в базе, но клан не найден, значит пользователь не в клане
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			// if debug_type {
			// 	log.Printf("Пользователь %d выходит из клана %d", user_id, clan_id)
			// }

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении владельца клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что-то пошло не так при выходе из клана..."))
				return
			}

			if role == "owner" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Владелец клана не может выйти из него"))
				return
			}

			creator_id_raw, err := helpers.GetClanOwnerID(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении владельца клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что-то пошло не так при выходе из клана..."))
				return
			}

			// 2. Выполняем выход из клана
			err = helpers.LeaveClan(clan_id, user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при выходе из клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что-то пошло не так при выходе из клана..."))
				return
			}

			creator_id := int64(creator_id_raw)

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы успешно вышли из клана"))
			bot.Send(tgbotapi.NewMessage(creator_id, fmt.Sprintf("Пользователь %s вышел из клана", message.From.FirstName)))
			return
		}
	case "getcode":
		{
			clan_id, err := helpers.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := helpers.GetUserClanRole(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			if role != "owner" && role != "admin" {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не владелец/админ клана"))
				return
			}

			invite_code, err := helpers.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Оишбка при получении приглашения"))
				return
			}
			if invite_code == "" {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет приглашения"))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваш код для приглашения: `%s`\nСкиньте его своим участникам: /clan invite %s", invite_code, invite_code)))
		}
	default:
		{
			formatted_actions := strings.Join(AvailableActions, ", ")
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Введите действие\nИспользуйте: /clan create [имя]\nВсе действия: "+formatted_actions))
			return
		}
	}
}
