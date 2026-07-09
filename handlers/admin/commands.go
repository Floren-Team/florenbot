package admin

import (
	helpers "florenbot/engine/helpers"
	"florenbot/engine/structs"
	std_helpers "florenbot/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	engine "florenbot/engine/mysql"
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
		text += fmt.Sprintf("🆔 **ID:** `%d`\n👤 **Название:** `%s`\nСистемное имя: %s\nКороткое имя: %s\n\n", role.ID, role.Name, role.BaseShort, role.ShortName)
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
	text := fmt.Sprintf("📋 **Профиль чата:** `%s`\n🆔 **ID:** `%d`\n\n", chat.Name, chat_id)

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
    db_id := uint64(chat_id)
    if chat_id < 0 {
        db_id = uint64(-chat_id)
    }

    // 1. Проверяем аргументы команды
    args := strings.Fields(message.CommandArguments())
    if len(args) < 2 {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат! Используйте: `/setrole [username] [role_id]`"))
        return
    }

    // 2. Ищем целевого пользователя
    username, err := std_helpers.ParseTelegramUsername(args[0])
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат юзернейма."))
        return
    }

    targetUserID := helpers.GetUserIDByUsername(username)
    if targetUserID == 0 {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Пользователь с таким username не найден."))
        return
    }

    // 3. Проверка прав администратора (ищем роль отправителя в текущем чате)
    var senderMember structs.Member
    err = engine.DB.Preload("Role").
        Where("user_id = ? AND chat_id = ?", message.From.ID, db_id).
        First(&senderMember).Error

    isAdmin := err == nil && (senderMember.Role.ShortName == "owner" || senderMember.Role.ShortName == "creator")
    if !isAdmin {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав администратора в этом чате."))
        return
    }

    // 4. Парсинг ID роли
    roleID, err := strconv.ParseUint(args[1], 10, 64)
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат ID роли."))
        return
    }

    // 5. Установка роли через обновленную функцию AddMemberRole (которую мы делали ранее)
    // Используем `baseShort` (ShortName) или меняем логику, чтобы принимать сразу ID роли
    err = helpers.SetMemberRole(targetUserID, roleID, db_id)
    if err != nil {
        log.Printf("Ошибка установки роли: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при сохранении роли в БД."))
        return
    }

    bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль пользователя %s успешно установлена.", username)))
}

// HandleNewRole — создание новой роли в чате
func HandleNewRole(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
	user_id := uint64(message.From.ID)
	chat_id := message.Chat.ID

	// 1. Получаем роль пользователя КОНКРЕТНО в этом чате
	memberRole, err := helpers.GetMemberRole(user_id, uint64(parsed_chat_id))
	if err != nil {
		log.Printf("DEBUG: [HandleNewRole] Ошибка получения роли в чате: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось проверить ваши права."))
		return
	}

	log.Printf("DEBUG: [HandleNewRole] Полученная роль: ID=%d, Name='%s'", memberRole.ID, memberRole.ShortName)

	// 2. Проверяем права (проверяем роль из таблицы members)
	// Убедись, что используешь правильное поле для названия роли (ShortName или Name)
	if !std_helpers.IsUserOwnerOrCreator(&memberRole) {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
		return
	}

	// 3. Проверяем аргументы команды
	args := strings.Fields(message.CommandArguments())
	if len(args) < 2 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Используйте: `/newrole [название_роли] [короткое_название_роли]`"))
		return
	}

	// 4. Создаем роль в базе данных
	if err := helpers.CreateRole(args[0], args[1], uint64(parsed_chat_id)); err != nil {
		log.Printf("DEBUG: [HandleNewRole] КРИТИЧЕСКАЯ ОШИБКА: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при работе с базой данных."))
		return
	}

	msg := tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль `%s` создана.", args[0]))
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

	// 2. Проверка аргументов (ожидаем: ID, новое имя, новое короткое имя)
	args := strings.Fields(message.CommandArguments())
	if len(args) < 3 {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Используйте: `/editrole [ID] [новое_имя] [новое_короткое_имя]`"))
		return
	}

	// Преобразуем ID
	roleID, _ := strconv.ParseUint(args[0], 10, 64)

	// Создаем объект для передачи в функцию
	roleToUpdate := structs.Role{
		ID:        roleID,
		ChatID:    uint64(parsed_chat_id),
		Name:      args[1],
		ShortName: args[2],
	}

	// 3. Вызов обновленной функции EditRole
	_, err = helpers.EditRole(roleToUpdate)
	if err != nil {
		log.Printf("DEBUG: [HandleEditRole] ОШИБКА: %v", err)
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка обновления в БД."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("✅ Роль с ID `%d` успешно изменена.", roleID)))
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
