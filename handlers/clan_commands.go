package handlers

import (
	"database/sql"
	"florenbot/engine"
	"florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

// HandleClan обрабатывает команду /clan
func HandleClan(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	parts := strings.Fields(args)
	var AvailableActions = []string{
		"create", "list", "delete", "update", "join",
		"leave", "createcode", "getcode", "invite",
		"delcode", "revoke", "kick", "add", "ban",
	}
	user_id := uint64(message.From.ID)
	debug_type := GetEnvBool("DEBUG", false)

	if len(parts) < 1 {
		clan_id, err := engine.GetUserClanID(user_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения клана: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
			return
		}

		clan, err := engine.GetClan(clan_id)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка получения клана: %v", err)
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клан не найден"))
			return
		}

		count, _ := engine.GetClanMemberCount(clan.Id)

		info := fmt.Sprintf("✅ *Информация о клане:*\n\n"+
			"🆔 ID: `%d`\n"+
			"🏷 Название: *%s*\n"+
			"👥 Участников: `%d`\n"+
			"👑 Имя владельца: `%s`",
			clan.Id, clan.Name, count, clan.OwnerName)

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
			_, err := engine.GetUserClanID(user_id)

			// 1. Проверяем: если ошибка есть, НО это не "нет строк" (sql.ErrNoRows),
			// значит произошла реальная ошибка БД (соединение, запрос и т.д.)
			if err != nil && err != sql.ErrNoRows {
				if debug_type {
					log.Printf("Ошибка получения клана из БД: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при обращении к базе данных"))
				return
			}

			// 2. Если ошибки нет (err == nil), значит клан найден,
			// пользователь уже состоит в клане
			if err == nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы уже состоите в клане"))
				return
			}

			user_id := uint64(message.From.ID)
			balance, err := engine.GetBalance(user_id, message.From.UserName)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения баланса: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не авторизованы"))
				return
			}

			if balance < 1000 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Недостаточно средств. Необходимо 1000 рублей"))
				return
			}

			// 3. Если мы дошли сюда, значит err == sql.ErrNoRows (клана нет).
			// Теперь проверяем формат команды
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат! Используйте: /clan create [имя]"))
				return
			}

			// Формируем полное имя пользователя
			fullName := message.From.FirstName
			if message.From.LastName != "" {
				fullName += " " + message.From.LastName
			}

			// Пытаемся создать клан
			name := strings.Join(parts[1:], " ")
			err = engine.CreateClan(name, fullName, user_id)

			// 4. Обязательно проверяем, удалось ли создать клан в БД
			if err != nil {
				if debug_type {
					log.Printf("Ошибка создания клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось создать клан"))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Клан создан: %s", name)))
			return
		}
	case "getrole":
		{
			_, err := engine.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			clan_id, err := engine.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			_, err = engine.GetUserClanID(user_reply_id_raw)
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
			if err := engine.KickClanUser(clan_id, user_reply_id_raw)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Пользователь %s кикнут из клана", reply.From.FirstName)))

			// Ответить пользователю
			msg := tgbotapi.NewMessage(user_reply_id, fmt.Sprintf("✅ Вы были кикнуты из клана: %s", reason))
			bot.Send(msg)
			return
		}
	case "list":
		{
			clans, err := engine.GetClans()
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
					clan.Id, clan.Name, clan.MemberCount,
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

			clan_id, err := engine.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			_, err = engine.GetUserClanID(user_reply_id)
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
			err = engine.BlockMemberClan(clan_id, user_reply_id, reason)

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
			_, err := engine.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			clan_id, err := engine.GetClanByOwnerID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка получения клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не являетесь владельцем клана"))
				return
			}

			err = engine.DeleteClan(clan_id)
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
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Клана с таким кодом не существует"))
				return
			}

			err = engine.CheckBlacklist(clan_id, user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вас заблокировали!"))
				return
			}

			err = engine.JoinClan(clan_id, user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при приглашении в клан: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Вы присоеденились в клан!"))

			// Отправляем владельцу сообщение о приглашении
			owner_id_raw, err := engine.GetClanOwnerID(clan_id)
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
				return // Додаємо return, щоб бот не виконував код далі
			}

			clan_id, err := engine.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			// Спочатку видаляємо код
			err = engine.RevokeInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}

			// Після видалення отримуємо новий код (або створюємо, якщо логіка передбачає)
			newCode, err := engine.GetClanInviteCode(clan_id)
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
			clan_id, err := engine.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			_, err = engine.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при удалении приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет кода приглашения"))
				return
			}

			err = engine.DeleteInviteCode(clan_id)
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

			clan_id, err := engine.GetUserClanID(user_id)
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

			existing_clan, _ := engine.GetUserClanID(target_user_id)
			if existing_clan != 0 {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь уже состоит в клане"))
				return
			}

			err = engine.AddUserToClan(clan_id, target_user_id)
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
			clan_id, err := engine.GetUserClanID(user_id)

			if err != nil {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			_, err = engine.GetClanInviteCode(clan_id)
			if err != nil && err != sql.ErrNoRows {
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
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
				if debug_type {
					log.Printf("Ошибка при создании приглашения: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что то пошло не так..."))
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Ваш код приглашения: `%s`, скиньте его своим участникам: /clan invite %s", invite_code, invite_code)))
		}
	case "leave":
		{
			// 1. Ищем клан, в котором состоит пользователь (как участник)
			clan_id, err := engine.GetUserClanID(user_id)
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

			role, err := engine.GetUserClanRole(user_id)
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

			creator_id_raw, err := engine.GetClanOwnerID(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении владельца клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Что-то пошло не так при выходе из клана..."))
				return
			}

			// 2. Выполняем выход из клана
			err = engine.LeaveClan(clan_id, user_id)
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
			clan_id, err := engine.GetUserClanID(user_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении клана: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Вы не состоите в клане"))
				return
			}

			role, err := engine.GetUserClanRole(user_id)
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

			invite_code, err := engine.GetClanInviteCode(clan_id)
			if err != nil {
				if debug_type {
					log.Printf("Ошибка при получении приглашения: %v", err)
				}
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
