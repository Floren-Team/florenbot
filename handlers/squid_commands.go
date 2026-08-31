package handlers

import (
	helpers "florenbot/engine/helpers"
	std_helpers "florenbot/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	structs "florenbot/engine/structs"
	engine "florenbot/engine/mysql"
	gorm "gorm.io/gorm"
	"errors"
	"fmt"
)

func HandleCreateRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

    // 1. Проверка прав
    memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
    if err != nil || !std_helpers.IsUserOwnerOrCreator(&memberRole) {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    userId := uint64(message.From.ID)

    roomToCreate := structs.SquidRooms{
        OwnerId: userId,
        Members: []structs.SquidMembers{
            {UserId: userId},
        },
    }

    err = helpers.CreateSquidRoom(engine.DB, roomToCreate)
    if err != nil {
        log.Printf("DEBUG: [HandleCreateRoom] Ошибка создания комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось создать комнату."))
        return
    }

    bot.Send(tgbotapi.NewMessage(chat_id, "Комната успешно создана!"))
}

func HandleSquidAll(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))
    user_id := message.From.ID

    exists, _, err := helpers.GetRoomSquid(engine.DB, uint64(user_id)) 
    if err != nil {
        log.Printf("DEBUG: [HandleSendMessageToAll] Ошибка получения комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось получить комнату."))
        return
    }

    if !exists {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет активной комнаты."))
        return
    }

    // 1. Проверка прав
    memberRole, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
    if err != nil || !std_helpers.IsUserOwnerOrCreator(&memberRole) {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    msg := tgbotapi.NewMessage(chat_id, "Внимание! `Игра в кальмара` собирает новых игроков\n"+
                                        "Призовой фонд: укажут в МП\n"+
                                        "Для участия в игре, напишите /joinsquid\n"+
                                        "Поторопитесь!")
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}

func HandleJoinRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))

    // 1. Проверка прав пользователя (может ли он вообще заходить / находиться в чате)
    _, err := helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Вы не можете посещать комнату."))
        return
    }

    // 2. Ищем активную открытую комнату, к которой можно присоединиться 
    // (например, последнюю созданную со статусом "open")
    var room structs.SquidRooms
    err = engine.DB.Where("status = ?", "open").Order("id DESC").First(&room).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Сейчас нет активных открытых комнат для входа."))
            return
        }
        log.Printf("DEBUG: [HandleJoinRoom] Ошибка поиска комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось найти комнату."))
        return
    }

    // 3. Добавляем пользователя в найденную комнату (передаем db, roomId, userId)
    err = helpers.JoinRoom(engine.DB, room.ID, uint64(message.From.ID))
    if err != nil {
        log.Printf("DEBUG: [HandleJoinRoom] Ошибка входа в комнату: %v", err)
        // Выводим текст ошибки из хелпера (например, "ви вже приєдналися до цієї кімнати")
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ "+err.Error()))
        return
    }

    msg := tgbotapi.NewMessage(chat_id, "Вы успешно вошли в комнату `Игра в кальмара`.")
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}

func HandleSquidInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    user_id := message.From.ID

    var room structs.SquidRooms
    // Виправляємо запит: підтягуємо і саму кімнату, і учасників, і профіль User для кожного учасника
    err := engine.DB.Preload("Members.User").Where("status = ?", "open").First(&room).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Сейчас нет активных открытых комнат."))
            return
        }
        log.Printf("DEBUG: [HandleSquidInfo] Ошибка получения комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось получить информацию о комнате."))
        return
    }

    // Формируем текст сообщения в Markdown
    text := fmt.Sprintf("🦑 *Информация о комнате Игры в кальмара*\n\n"+
        "🆔 ID комнаты: `%d`\n"+
        "📌 Статус: `%s`\n"+
        "👥 Участников: `%d`\n\n",
        room.ID, room.Status, len(room.Members))

    isMember := false

    if len(room.Members) > 0 {
        text += "*Список участников:*\n"
        for i, member := range room.Members {
            if member.UserId == uint64(user_id) {
                isMember = true
            }

            name := member.User.FirstName
            if name == "" {
                name = fmt.Sprintf("ID: %d", member.UserId)
            }
            text += fmt.Sprintf("%d. %s\n", i+1, name)
        }
    } else {
        text += "_Пока нет участников._"
    }

    msg := tgbotapi.NewMessage(chat_id, text)
    msg.ParseMode = "Markdown"

    // Динамічна кнопка: якщо користувач уже в кімнаті — показуємо "Вийти", якщо ні — "Войти"
    var keyboard tgbotapi.InlineKeyboardMarkup
    if isMember {
        keyboard = tgbotapi.NewInlineKeyboardMarkup(
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonData("🚪 Выйти", "squid_leave"),
            ),
        )
    } else {
        keyboard = tgbotapi.NewInlineKeyboardMarkup(
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonData("📥 Войти", "squid_join"),
            ),
        )
    }
    msg.ReplyMarkup = keyboard

    bot.Send(msg)
}


func HandleSquidCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
    user_id := callback.From.ID
    data := callback.Data

    switch data {
    case "squid_join":
        var room structs.SquidRooms
        err := engine.DB.Where("status = ?", "open").Order("id DESC").First(&room).Error
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                answerCallback(bot, callback.ID, "❌ Нет активных комнат.")
                return
            }
            log.Printf("DEBUG: [HandleSquidCallback] Ошибка поиска комнаты: %v", err)
            answerCallback(bot, callback.ID, "❌ Ошибка при поиске комнаты.")
            return
        }

        err = helpers.JoinRoom(engine.DB, room.ID, uint64(user_id))
        if err != nil {
            answerCallback(bot, callback.ID, "❌ "+err.Error())
            return
        }

        answerCallback(bot, callback.ID, "✅ Вы успешно вошли в комнату!")

    case "squid_leave":
        var member structs.SquidMembers
        err := engine.DB.Where("user_id = ?", uint64(user_id)).First(&member).Error
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                answerCallback(bot, callback.ID, "❌ Вы не состоите ни в одной комнате.")
                return
            }
            log.Printf("DEBUG: [HandleSquidCallback] Ошибка поиска участника: %v", err)
            answerCallback(bot, callback.ID, "❌ Ошибка.")
            return
        }

        err = helpers.LeaveRoom(engine.DB, member.RoomId, uint64(user_id))
        if err != nil {
            answerCallback(bot, callback.ID, "❌ "+err.Error())
            return
        }

        answerCallback(bot, callback.ID, "✅ Вы вышли из комнаты.")

    default:
        answerCallback(bot, callback.ID, "❓ Неизвестное действие.")
    }
}


// Вспомогательная функция для ответа на callback-запрос (убирает крутилку загрузки у кнопки)
func answerCallback(bot *tgbotapi.BotAPI, callbackID string, text string) {
    callbackCfg := tgbotapi.NewCallback(callbackID, text)
    if _, err := bot.Request(callbackCfg); err != nil {
        log.Printf("DEBUG: [AnswerCallback] Ошибка: %v", err)
    }
}

func HandleLeaveRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))
    user_id := message.From.ID

    // 1. Проверка прав
    _, err := helpers.GetMemberRole(uint64(user_id), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    // Получаем комнату пользователя
    exists, room, err := helpers.GetRoomSquid(engine.DB, uint64(user_id)) 
    if err != nil {
        log.Printf("DEBUG: [HandleLeaveRoom] Ошибка получения комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось получить комнату."))
        return
    }

    if !exists {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет активной комнаты."))
        return
    }

    // Выходим из комнаты
    err = helpers.LeaveRoom(engine.DB, room.ID, uint64(user_id))
    if err != nil {
        log.Printf("DEBUG: [HandleLeaveRoom] Ошибка выхода из комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("❌ Не удалось выйти из комнаты. %v", err)))
        return
    }

    msg := tgbotapi.NewMessage(chat_id, "Вы вышли из комнаты `Игра в кальмара`.")
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}