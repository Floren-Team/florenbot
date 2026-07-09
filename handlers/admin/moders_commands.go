package admin

import (
    "strings"
	"time"
	"log"
    helpers "florenbot/engine/helpers"
    std_helpers "florenbot/helpers"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBan обрабатывает команду /ban, поддерживая ответы (reply) и указание @username
func HandleBan(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    var targetUserID int64
    var reason string
    args := strings.Fields(message.CommandArguments())

    // 1. Определяем, кого банить
    if message.ReplyToMessage != nil {
        // Если есть ответ на сообщение, берем ID того, кому отвечаем
        targetUserID = message.ReplyToMessage.From.ID
        // Причина — это весь текст после команды /ban
        reason = strings.Join(args, " ")
    } else {
        // Если ответа нет, проверяем аргументы
        if len(args) < 1 {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: укажите пользователя через @username или ответом на сообщение."))
            return
        }

        // Парсим юзернейм
        username, err := std_helpers.ParseTelegramUsername(args[0])
        if err != nil {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат юзернейма."))
            return
        }

        // Получаем ID пользователя из вашей базы данных по юзернейму
        targetUserID = int64(helpers.GetUserIDByUsername(username))
        if targetUserID == 0 {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь с таким username не найден в базе."))
            return
        }

        // Причина — это все аргументы, начиная со второго
        if len(args) > 1 {
            reason = strings.Join(args[1:], " ")
        }
    }

    // 2. Получаем информацию о пользователе из БД для подтверждения
    userProfile, err := helpers.GetUserByID(uint64(targetUserID))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден в базе данных."))
        return
    }

    // 3. Выполняем бан через Telegram API
    banConfig := tgbotapi.BanChatMemberConfig{
        ChatMemberConfig: tgbotapi.ChatMemberConfig{
            ChatID: message.Chat.ID,
            UserID: targetUserID,
        },
    }
    
    _, err = bot.Request(banConfig)
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось забанить: "+err.Error()))
        return
    }

    // 4. Отправляем подтверждение
    response := "✅ Пользователь " + userProfile.Username + " забанен."
    if reason != "" {
        response += "\nПричина: " + reason
    }
    
    bot.Send(tgbotapi.NewMessage(message.Chat.ID, response))
}


func HandleKick(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    var targetUserID int64
    var reason string
    args := strings.Fields(message.CommandArguments())

    // 1. Определяем, кого кикнуть
    if message.ReplyToMessage != nil {
        targetUserID = message.ReplyToMessage.From.ID
        reason = strings.Join(args, " ")
    } else {
        if len(args) < 1 {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: укажите пользователя через @username или ответом на сообщение."))
            return
        }

        username, err := std_helpers.ParseTelegramUsername(args[0])
        if err != nil {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат юзернейма."))
            return
        }

        targetUserID = int64(helpers.GetUserIDByUsername(username))
        if targetUserID == 0 {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден в базе."))
            return
        }

        if len(args) > 1 {
            reason = strings.Join(args[1:], " ")
        }
    }

    // 2. Получаем информацию о пользователе
    userProfile, err := helpers.GetUserByID(uint64(targetUserID))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден в базе данных."))
        return
    }

    // 3. Выполняем "Kick": Бан + Сразу разбан (чтобы пользователь мог вернуться)
    // Шаг А: Баним пользователя
    _, err = bot.Request(tgbotapi.BanChatMemberConfig{
        ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: message.Chat.ID, UserID: targetUserID},
    })
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось кикнуть: "+err.Error()))
        return
    }

    // Шаг Б: Сразу разбаниваем, чтобы снять ограничение
    _, _ = bot.Request(tgbotapi.UnbanChatMemberConfig{
        ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: message.Chat.ID, UserID: targetUserID},
        OnlyIfBanned:     true,
    })

    // 4. Подтверждение
    response := "✅ Пользователь " + userProfile.Username + " был кикнут."
    if reason != "" {
        response += "\nПричина: " + reason
    }
    
    bot.Send(tgbotapi.NewMessage(message.Chat.ID, response))
}


// HandleMute обрабатывает команду /mute @user <duration> <reason>
func HandleMute(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	var targetUserID int64
	var durationStr string
	var reason string
	args := strings.Fields(message.CommandArguments())

	// 1. Определение пользователя
	if message.ReplyToMessage != nil {
		targetUserID = message.ReplyToMessage.From.ID
		if len(args) < 1 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: укажите длительность (например, 1h, 30m)."))
			return
		}
		durationStr = args[0]
		reason = strings.Join(args[1:], " ")
	} else {
		if len(args) < 2 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: используйте /mute @user <duration> <reason>"))
			return
		}
		
		username, err := std_helpers.ParseTelegramUsername(args[0])
		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат юзернейма."))
			return
		}
		targetUserID = int64(helpers.GetUserIDByUsername(username))
		durationStr = args[1]
		reason = strings.Join(args[2:], " ")
	}

	// 2. Парсинг длительности
	duration, err := std_helpers.ParseDuration(durationStr)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат времени (используйте 30m, 1h, 1d)."))
		return
	}

	// 3. Выполнение мута
	// UntilDate — это время окончания мута в формате Unix Timestamp
	untilDate := time.Now().Add(duration).Unix()
	
	muteConfig := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: message.Chat.ID,
			UserID: targetUserID,
		},
		UntilDate: untilDate,
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages: false, // Запрещаем отправку сообщений
		},
	}

	_, err = bot.Request(muteConfig)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось ограничить пользователя: "+err.Error()))
		return
	}

	// 4. Подтверждение
    userProfile, err := helpers.GetUserByID(uint64(targetUserID))
    
    // Вместо проверки на nil, проверяем err
    username := "Пользователь"
    if err == nil {
        // Если ошибки нет, значит юзер найден, используем его имя
        username = userProfile.Username
    }

	response := "🔇 " + username + " отправлен в мут на " + durationStr + "."
	if reason != "" {
		response += "\nПричина: " + reason
	}
	
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, response))
}


// HandleUnMute обрабатывает команду /unmute @user
func HandleUnMute(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    var targetUserID int64
    args := strings.Fields(message.CommandArguments())

    // 1. Определение пользователя (Reply или @username)
    if message.ReplyToMessage != nil {
        targetUserID = message.ReplyToMessage.From.ID
    } else {
        if len(args) < 1 {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка: укажите пользователя через @username или ответом на сообщение."))
            return
        }

        username, err := std_helpers.ParseTelegramUsername(args[0])
        if err != nil {
            bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат юзернейма."))
            return
        }
        targetUserID = int64(helpers.GetUserIDByUsername(username))
    }

    if targetUserID == 0 {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Пользователь не найден."))
        return
    }

    // 2. Выполнение размута
    // Для снятия мута нужно разрешить отправку сообщений (CanSendMessages: true)
   unmuteConfig := tgbotapi.RestrictChatMemberConfig{
        ChatMemberConfig: tgbotapi.ChatMemberConfig{
            ChatID: message.Chat.ID,
            UserID: targetUserID,
        },
        Permissions: &tgbotapi.ChatPermissions{
            CanSendMessages:      true,
            CanSendMediaMessages: true,
            CanSendOtherMessages: true,
            CanAddWebPagePreviews: true,
        },
    }

    _, err := bot.Request(unmuteConfig)
    if err != nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось снять мут: "+err.Error()))
        return
    }

    // 3. Подтверждение
    userProfile, err := helpers.GetUserByID(uint64(targetUserID))
    username := "Пользователь"
    if err == nil {
        username = userProfile.Username
    }

    bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ С пользователя "+username+" сняты ограничения."))
}


// HandleDeleteMessage обрабатывает команду /del
func HandleDeleteMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    userID := uint64(message.From.ID)

    // Получаем пользователя
    _, err := helpers.GetUserByID(userID)
    if err != nil {
        log.Printf("Ошибка получения юзера: %v", err)
        return
    }

    // Получаем чат (исправлено получение ID для групп)
    chatID := message.Chat.ID
    searchID := chatID
    if searchID < 0 {
        searchID = -searchID
    }
    chat, err := helpers.GetChatById(searchID)
    if err != nil || chat == nil {
        log.Printf("Ошибка получения чата: %v", err)
        bot.Send(tgbotapi.NewMessage(chatID, "❌ Чат не найден."))
        return
    }

    // Проверка прав (теперь проверяем по ID роли или имени из структуры Role)
    // Убедись, что user.Role не nil перед проверкой
    // isAdmin := user.Role != nil && (user.Role.Name == "admin" || user.Role.Name == "owner" || user.Role.Name == "creator")
    // if !isAdmin {
    //     bot.Send(tgbotapi.NewMessage(chatID, "❌ У вас нет прав для выполнения этой команды."))
    //     return
    // }

    // Команда должна быть ответом на другое сообщение
    if message.ReplyToMessage == nil {
        bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка: используйте эту команду в ответ на сообщение, которое нужно удалить."))
        return
    }

    // 1. Удаляем сообщение, на которое ответили
    _, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, message.ReplyToMessage.MessageID))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось удалить сообщение: "+err.Error()))
        return
    }

    // 2. Удаляем саму команду /del
    _, _ = bot.Request(tgbotapi.NewDeleteMessage(chatID, message.MessageID))
}