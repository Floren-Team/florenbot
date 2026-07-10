package admin

import (
	helpers "florenbot/engine/helpers"
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	std_helpers "florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func HandleRoles(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID
	parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

	// 1. Получаем данные чата
	chat, err := helpers.GetChatById(int64(parsed_chat_id))
	if err != nil {
		log.Printf("Ошибка получения чата: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Чат не найден."))
		return
	}

	// 2. Получаем список всех ролей в чате
	roles, err := helpers.GetRolesByChatID(uint64(parsed_chat_id))
	if err != nil {
		log.Printf("Ошибка получения ролей: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка загрузки ролей."))
		return
	}

	// 3. Формируем текст сообщения

	// Заголовок сообщения
	text := fmt.Sprintf("📋 **Профиль чата:** `%s`\n\n", chat.Name)

	// Список ролей
	for _, role := range roles {
		text += fmt.Sprintf("🆔 **ID:** `%d`\n👤 **Название:** `%s`\nПриоритет: %d\nСистемное имя: %s\nКороткое имя: %s\n\n", role.ID, role.Name, role.Priority, role.BaseShort, role.ShortName)
	}
	msg := tgbotapi.NewMessage(chat_id, text)
	msg.ParseMode = "Markdown"

	// Отправляем сообщение
	bot.Send(msg)
}

func HandleStaff(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID

	// Преобразование ID для корректной работы с базой данных
	searchId := uint64(chat_id)
	if searchId > 0x8000000000000000 {
		searchId = uint64(-chat_id)
	}

	// 1. Получаем данные чата
	chat, err := helpers.GetChatById(int64(searchId))
	if err != nil {
		log.Printf("Ошибка получения чата: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Чат не найден."))
		return
	}

	// 2. Получаем список всех ролей в чате
	roles, err := helpers.GetRolesByChatID(searchId)
	if err != nil {
		log.Printf("Ошибка получения ролей: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка загрузки ролей."))
		return
	}

	// Заголовок сообщения
	text := fmt.Sprintf("📋 **Профиль чата:** `%s`\n\n", chat.Name)

	// 3. Формируем списки пользователей для каждой роли
	if len(roles) > 0 {
		for _, role := range roles {
			text += fmt.Sprintf("🎭 **Роль:** `%s` (ID: **`%d`**)\n", role.Name, role.ID)

			// Получаем список пользователей для текущей роли
			users, err := helpers.GetUsersByRole(role.ID, searchId)

			if err == nil && len(users) > 0 {
				for _, user := range users {
					// Выбираем имя: используем username, если он есть, иначе берем FirstName
					displayName := user.Username
					if displayName == "" {
						displayName = user.FirstName
					}

					// Форматирование: добавляем @ для никнеймов, для имен — без
					if user.Username != "" {
						text += fmt.Sprintf("└ @%s\n", displayName)
					} else {
						text += fmt.Sprintf("└ %s\n", displayName)
					}
				}
			} else {
				text += "└ _Пользователей нет_\n"
			}
			text += "\n" // Пустая строка между блоками ролей для читаемости
		}
	} else {
		text += "_Роли в этом чате еще не настроены._"
	}

	// 4. Отправка итогового сообщения
	msg := tgbotapi.NewMessage(chat_id, text)
	msg.ParseMode = "Markdown" // Используем Markdown для красивого отображения
	bot.Send(msg)
}

// HandleSetRole обрабатывает команду /setrole <username> <role_id>
func HandleSetRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    db_id := std_helpers.ParseChatID(uint64(chat_id))

    args := strings.Fields(message.CommandArguments())
    
    var targetUserID uint64
    var targetUsername string
	var targetName string
    var roleID uint64
    var err error

    // 1. Определение цели (Reply или Username)
    if message.ReplyToMessage != nil {
        targetUserID = uint64(message.ReplyToMessage.From.ID)
        targetUsername = message.ReplyToMessage.From.UserName
		targetName = message.ReplyToMessage.From.FirstName
        
        if len(args) < 1 {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат! Используйте: `/setrole [role_id]` при ответе на сообщение."))
            return
        }
        
        roleID, err = strconv.ParseUint(args[0], 10, 64)
    } else if len(args) >= 2 {
        targetUsername, err = std_helpers.ParseTelegramUsername(args[0])
        if err != nil {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат юзернейма."))
            return
        }
        targetUserID = helpers.GetUserIDByUsername(targetUsername)
        if targetUserID == 0 {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Пользователь с таким username не найден."))
            return
        }
        roleID, err = strconv.ParseUint(args[1], 10, 64)
    } else {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Формат: `/setrole [username] [role_id]` или ответьте на сообщение `/setrole [role_id]`"))
        return
    }

    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат ID роли."))
        return
    }

    // 2. Проверка прав администратора
    var senderMember structs.Member
    err = engine.DB.Preload("Role").
        Where("user_id = ? AND chat_id = ?", message.From.ID, db_id).
        First(&senderMember).Error

    isAdmin := err == nil && (senderMember.Role.ShortName == "owner" || senderMember.Role.ShortName == "creator")
    if !isAdmin {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав администратора в этом чате."))
        return
    }

	_, err = helpers.GetRoleByID(roleID, uint64(db_id))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Роль не найдена."))
		return
	}

    // 3. Установка роли
    err = helpers.SetMemberRole(targetUserID, roleID, uint64(db_id))
    if err != nil {
        log.Printf("Ошибка установки роли: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при сохранении роли в БД."))
        return
    }

    // Формируем красивый вывод, если юзернейм пустой (например, если нет @ в профиле)
    displayName := targetUsername
    if displayName == "" {
        displayName = targetName // Теперь переменная используется!
    }
    if displayName == "" {
        displayName = fmt.Sprintf("%d", targetUserID)
    }
    bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль пользователя @%s успешно установлена.", displayName)))
}
// HandleNewRole — создание новой роли в чате
func HandleNewRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
	user_id := uint64(message.From.ID)
	chat_id := message.Chat.ID

	memberRole, err := helpers.GetMemberRole(user_id, uint64(parsed_chat_id))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось проверить ваши права."))
		return
	}

	if !std_helpers.IsUserOwnerOrCreator(&memberRole) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	// 3. Проверяем аргументы команды
	// Ожидаемый формат: /newrole [название] [короткое_имя] [служебное_имя] [приоритет]
	args := strings.Fields(message.CommandArguments())
	if len(args) < 4 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Используйте: `/newrole [название] [короткое_имя] [служебное_имя] [приоритет]`"))
		return
	}

	// Парсим приоритет из последнего аргумента
	priority, err := strconv.Atoi(args[len(args)-1])
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Приоритет должен быть числом."))
		return
	}

	if priority < 10 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Приоритет должен быть больше 10."))
		return
	}

	if priority > 100 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Приоритет должен быть меньше 100."))
		return
	}

	base_short := args[len(args)-2]
	role_short_name := args[len(args)-3]
	role_name := strings.Join(args[:len(args)-3], " ")

	// Проверка роли

	_, err = helpers.GetRoleByShortName(role_short_name, uint64(parsed_chat_id))
	if err == nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Роль с таким названием уже существует."))
		return
	}

	switch base_short {
	case "admin", "creator", "owner", "moderator", "member":
		// Вызываем обновленную функцию CreateRole с приоритетом
		if err := helpers.CreateRole(role_name, role_short_name, base_short, uint64(parsed_chat_id), priority); err != nil {
			log.Printf("DEBUG: [HandleNewRole] ОШИБКА БД: %v", err)
			bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при создании роли в БД."))
			return
		}
	default:
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверное служебное имя. Доступные: admin, creator, owner, moderator, member"))
		return
	}

	msg := tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль `%s` создана (Приоритет: %d).", role_name, priority))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleModers(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID
	parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

	// 1. Проверка прав (объединил условия для чистоты кода)
	memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil || (!std_helpers.IsUserAdmin(&memberRole) && !std_helpers.IsUserOwnerOrCreator(&memberRole)) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	// 2. Получаем список модераторов
	moderators, err := helpers.GetModeratorsUsers(uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleModers] ОШИБКА БД: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при получении списка модераторов."))
		return
	}

	// 3. Отправляем список
	msgText := "👮‍♂️ Список модераторов:"
	if len(moderators) == 0 {
		msgText += "\nСписок пуст."
	} else {
		for _, m := range moderators {
			msgText += fmt.Sprintf("\n- %s (@%s)", m.FirstName, m.Username)
		}
	}

	msg := tgbotapi.NewMessage(chat_id, msgText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleAdmins(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID
	parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

	// 1. Проверка прав
	memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil || (!std_helpers.IsUserAdmin(&memberRole) && !std_helpers.IsUserOwnerOrCreator(&memberRole)) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	// 2. Получаем список админов
	admins, err := helpers.GetAdminsUsers(uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleAdmins] ОШИБКА БД: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при получении списка админов."))
		return
	}

	// 3. Отправляем список
	msgText := "👮‍♀️ Список администраторов:"
	if len(admins) == 0 {
		msgText += "\nСписок пуст."
	} else {
		for _, a := range admins {
			msgText += fmt.Sprintf("\n- %s (@%s)", a.FirstName, a.Username)
		}
	}

	msg := tgbotapi.NewMessage(chat_id, msgText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// HandleEditRole — редактирование роли в чате

func HandleEditRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
	chat_id := message.Chat.ID

	// 1. Проверка прав пользователя
	memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil || !std_helpers.IsUserOwnerOrCreator(&memberRole) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав или ошибка проверки."))
		return
	}

	// 2. Проверка аргументов
	args := strings.Fields(message.CommandArguments())
	// Ожидаем: ID (1) + Имя (N слов) + Короткое имя (1 слово) = минимум 3 аргумента
	if len(args) < 3 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Используйте: `/editrole [ID] [новое_название] [новое_короткое_имя]`"))
		return
	}

	// ID всегда первый аргумент
	roleID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Некорректный ID роли."))
		return
	}

	// Короткое имя — последний аргумент
	new_short_name := args[len(args)-1]

	// Название — всё, что между ID и коротким именем
	new_name := strings.Join(args[1:len(args)-1], " ")

	// Создаем объект для передачи в функцию
	roleToUpdate := structs.Role{
		ID:        roleID,
		ChatID:    uint64(parsed_chat_id),
		Name:      new_name,
		ShortName: new_short_name,
	}

	// 3. Вызов функции EditRole
	_, err = helpers.EditRole(roleToUpdate)
	if err != nil {
		log.Printf("DEBUG: [HandleEditRole] ОШИБКА: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка обновления в БД."))
		return
	}
	msg := tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль `%s` (ID `%d`) успешно изменена.", new_name, roleID))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// HandleDeleteRole — удаление роли из чата по её ID
func HandleDeleteRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
	chat_id := message.Chat.ID

	// 1. Проверка существования чата
	if _, err := helpers.GetChatById(parsed_chat_id); err != nil {
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ Чат не найден."))
		return
	}

	// 2. Проверка прав пользователя
	memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleDeleteRole] Ошибка получения роли в чате: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось проверить ваши права."))
		return
	}

	if !std_helpers.IsUserOwnerOrCreator(&memberRole) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	// 3. Проверка аргументов
	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ Неверный формат! Используйте: `/delrole [ID_роли]`"))
		return
	}

	// Конвертируем строку аргумента в числовой ID
	roleID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ ID роли должен быть числом."))
		return
	}

	role, err := helpers.GetRoleByID(roleID, uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleDeleteRole] ОШИБКА: %v", err)
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ Роли с таким ID не существует."))
		return
	}

	if role.BaseShort == "member" || role.BaseShort == "owner" || role.BaseShort == "creator" {
		msg := tgbotapi.NewMessage(int64(chat_id), fmt.Sprintf("❌ Нельзя удалить роль `%s`, так как она является базовой.", role.ShortName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// Проверяем если у роли есть пользователи

	members, err := helpers.GetUsersByRole(roleID, uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleDeleteRole] ОШИБКА: %v", err)
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ При удалении роли произошла ошибка."))
		return
	}

	if len(members) > 0 {
		msg := tgbotapi.NewMessage(int64(chat_id), fmt.Sprintf("❌ Роль `%s` не может быть удалена, так как у нее есть пользователи.\n"+
			"Снимите с них роли (/rr) чтобы удалить роль.", role.ShortName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// 4. Удаление роли
	if err := helpers.DeleteRoleByID(roleID, uint64(parsed_chat_id)); err != nil {
		log.Printf("Ошибка удаления роли: %v", err)
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ Ошибка при удалении: возможно, такой роли не существует."))
		return
	}

	msg := tgbotapi.NewMessage(int64(chat_id), fmt.Sprintf("✅ Роль с ID `%d` успешно удалена.", roleID))
	msg.ParseMode = "Markdown"

	bot.Send(msg)
}

// Общая функция для добавления роли с проверкой иерархии
func addRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message, shortName string) {
    parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
    reply := message.ReplyToMessage

    // 1. Получаем роль цели
    targetRole, err := helpers.GetMemberRole(uint64(reply.From.ID), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить роль пользователя."))
        return
    }

    // 2. ЗАЩИТА ИЕРАРХИИ
    // Нельзя повышать/понижать того, кто выше или равен по приоритету
    // (кроме случаев, когда мы просто меняем роль, но тут мы проверяем безопасность)
    if targetRole.Priority >= 50 && shortName == "moderator" { // Например, если цель уже админ(50+)
         bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Нельзя назначить модератором того, кто уже является администратором или выше."))
         return
    }
    
    // Если пытаемся назначить админом того, кто уже владелец(90) или создатель(100)
    if targetRole.Priority >= 90 && shortName == "admin" {
         bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Нельзя назначить администратором владельца или создателя."))
         return
    }

    // 3. Получение новой роли и сохранение
    role, err := helpers.GetRoleByShortName(shortName, uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Роль не найдена."))
        return
    }

    err = helpers.SetMemberRole(uint64(reply.From.ID), role.ID, uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка БД."))
        return
    }

    msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Пользователь `%s` теперь `%s`.", reply.From.FirstName, shortName))
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}

func HandleAddModer(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(std_helpers.ParseChatID(uint64(message.Chat.ID))))
    if err != nil || !std_helpers.IsUserAdmin(&memberRole) {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет прав."))
        return
    }
    if message.ReplyToMessage == nil { return }
    addRole(bot, message, "moderator")
}

func HandleAddAdmin(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(std_helpers.ParseChatID(uint64(message.Chat.ID))))
    if err != nil || !std_helpers.IsUserOwnerOrCreator(&memberRole) {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ У вас нет прав."))
        return
    }
    if message.ReplyToMessage == nil { return }
    addRole(bot, message, "admin")
}

func HandleAddOwner(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

    // 1. Проверка прав (только Владелец может добавлять другого Владельца)
    memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
    if err != nil || !std_helpers.IsUserOwner(&memberRole) {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    // 2. Проверка reply
    if message.ReplyToMessage == nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Команда должна быть ответом на сообщение."))
        return
    }
    
    reply_id := uint64(message.ReplyToMessage.From.ID)

    // 3. Проверка лимита владельцев
    role, err := helpers.GetRoleByShortName("owner", uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Роль 'owner' не найдена."))
        return
    }

    users, err := helpers.GetUsersByRole(role.ID, uint64(parsed_chat_id))
    if err == nil && len(users) >= 2 {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Нельзя добавить больше двух владельцев. Лимит достигнут."))
        return
    }

    // 4. Дополнительная защита: нельзя назначить владельцем того, кто уже выше или равен
    targetRole, _ := helpers.GetMemberRole(reply_id, uint64(parsed_chat_id))
    if targetRole.Priority >= 90 {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Пользователь уже имеет равный или более высокий статус."))
        return
    }

    // 5. Выполнение добавления через общую функцию
    addRole(bot, message, "owner")
}

func HandlePin(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID
	parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))
	// 1. Проверка существования чата
	if _, err := helpers.GetChatById(std_helpers.ParseChatID(uint64(parsed_chat_id))); err != nil {
		bot.Send(tgbotapi.NewMessage(int64(chat_id), "❌ Чат не найден."))
		return
	}

	// 2. Проверка прав пользователя
	memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandlePin] Ошибка получения роли в чате: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось проверить ваши права."))
		return
	}

	if !std_helpers.IsUserOwnerOrCreator(&memberRole) && !std_helpers.IsUserAdmin(&memberRole) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	reply := message.ReplyToMessage

	if reply == nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Команда должна быть ответом на сообщение."))
		return
	}

	reply_id := reply.MessageID

	pinConfig := tgbotapi.PinChatMessageConfig{
		ChatID:    chat_id,
		MessageID: reply_id,
	}

	// Пытаемся закрепить сообщение
	_, err = bot.Request(pinConfig)
	if err != nil {
		log.Printf("DEBUG: [HandlePin] Ошибка закрепления: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось закрепить сообщение."))
		return
	}

	msg := tgbotapi.NewMessage(chat_id, "✅ Сообщение успешно закреплено.")
	bot.Send(msg)
}

func HandleRestrictUserCmd(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chat_id := message.Chat.ID

	// 1. Проверка: является ли сообщение ответом на другое сообщение
	reply := message.ReplyToMessage
	if reply == nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Нужно ответить на сообщение пользователя."))
		return
	}

	// 2. Парсинг аргументов: /restrict [add/remove] [command]
	args := strings.Fields(message.CommandArguments())
	if len(args) < 2 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Использование: `/restrict [add|remove] [команда]`"))
		return
	}

	action := strings.ToLower(args[0])
	command := strings.TrimPrefix(args[1], "/") // Убираем /, если пользователь написал /ban

	userID := uint64(reply.From.ID)
	chatID := uint64(chat_id)

	// 3. Выполнение действия
	switch action {
	case "add":
		// Проверяем, установлено ли уже такое ограничение
		if helpers.IsCommandRestricted(userID, chatID, command) {
			bot.Send(tgbotapi.NewMessage(chat_id, "⚠️ Это ограничение уже установлено."))
			return
		}

		restriction := structs.UserCommandRestriction{
			UserID:  userID,
			ChatID:  chatID,
			Command: command,
		}
		// Сохраняем новое ограничение в базу данных
		if err := engine.DB.Create(&restriction).Error; err != nil {
			bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при сохранении ограничения."))
			return
		}
		bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Команда /%s запрещена для пользователя %s.", command, reply.From.FirstName)))

	case "remove":
		// Удаляем ограничение из базы данных
		result := engine.DB.Where("user_id = ? AND chat_id = ? AND command = ?", userID, chatID, command).
			Delete(&structs.UserCommandRestriction{})

		if result.RowsAffected == 0 {
			bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ограничение не найдено."))
		} else {
			bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Ограничение на команду /%s снято.", command)))
		}
	default:
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неизвестное действие. Используйте `add` или `remove`."))
	}
}

func HandleRemoveRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

    // 1. Получение ролей
    adminRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при проверке ваших прав."))
        return
    }

    reply := message.ReplyToMessage
    if reply == nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Команда должна быть ответом на сообщение."))
        return
    }

    targetRole, err := helpers.GetMemberRole(uint64(reply.From.ID), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось получить роль пользователя."))
        return
    }

    // 2. Защита от снятия роли самому себе
    if reply.From.ID == message.From.ID {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Нельзя снять роль самому себе."))
        return
    }

    // 3. Дебаг-логи для проверки приоритетов
    log.Printf("[DEBUG] --- [HandleRemoveRole] ---")
    log.Printf("[DEBUG] Админ: %s (ID: %d), Приоритет: %d", message.From.FirstName, message.From.ID, adminRole.Priority)
    log.Printf("[DEBUG] Цель: %s (ID: %d), Приоритет: %d", reply.From.FirstName, reply.From.ID, targetRole.Priority)


	if targetRole.BaseShort == "member" {
		msg := tgbotapi.NewMessage(chat_id, "❌ Пользователь не имеет роли. Снятие роли невозможно.")
		msg.ParseMode = "Markdown"
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
		return
	}

    // 4. ЗАЩИТА: Сравнение приоритетов
    // Админ может снять роль только в том случае, если его приоритет СТРОГО ВЫШЕ приоритета цели.
    // Если приоритет админа равен или ниже приоритета цели — действие запрещено.
    if adminRole.Priority <= targetRole.Priority {
        log.Printf("[DEBUG] ОТКАЗ: %d (админ) <= %d (цель)", adminRole.Priority, targetRole.Priority)
		msg := tgbotapi.NewMessage(chat_id, "❌ Ошибка: вы не можете снять роль у пользователя с равным или более высоким рангом.")
		msg.ParseMode = "Markdown"
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
        return
    }

    // 5. Удаление роли
    log.Printf("[DEBUG] Приоритет позволяет действие. Выполняем удаление...")
    err = helpers.RemoveRole(uint64(reply.From.ID), uint64(parsed_chat_id))
    if err != nil {
        log.Printf("[DEBUG] Ошибка БД: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при обновлении роли в БД."))
        return
    }

    log.Printf("[DEBUG] Роль успешно удалена для %s", reply.From.FirstName)

    // 6. Отправка результата

    text := fmt.Sprintf("✅ Пользователь **%s** снят с роли `%s`", reply.From.FirstName, targetRole.ShortName)
    msg := tgbotapi.NewMessage(chat_id, text)
    msg.ParseMode = "Markdown"
	msg.ReplyToMessageID = message.MessageID

    _, sendErr := bot.Send(msg)
    if sendErr != nil {
        log.Printf("[DEBUG] Ошибка отправки сообщения: %v", sendErr)
    }
}