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
	"time"
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


	if room.Status != "open" {
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
	
}

func HandleStartSquid(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))
    user_id := message.From.ID

    // 1. Проверяем права пользователя
    memberRole, err := helpers.GetMemberRole(uint64(user_id), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    var room structs.SquidRooms
    var queryErr error

    // Если создатель/админ чата — берем последнюю комнату, если обычный — его собственную
    if std_helpers.IsUserOwnerOrCreator(&memberRole) {
        queryErr = engine.DB.Order("id DESC").First(&room).Error
    } else {
        queryErr = engine.DB.Where("owner_id = ?", uint64(user_id)).Order("id DESC").First(&room).Error
    }

    if queryErr != nil {
        if errors.Is(queryErr, gorm.ErrRecordNotFound) {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Комната не найдена."))
            return
        }
        log.Printf("DEBUG: [HandleStartSquid] Ошибка поиска комнаты: %v", queryErr)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось найти комнату."))
        return
    }

    // 2. Проверка статуса: игра не может быть начатой, если комната открыта
    if room.Status != "closed" {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Игра не может быть начатой, если комната открыта."))
        return
    }

    // 3. Запускаем первую игру, передавая реальный ID найденной комнаты
    HandleStartFirstGame(bot, chat_id, room.ID)
}

// HandleStartFirstGame запускает первую игру "Красный свет, зеленый свет", таймер и автоматическое переключение света
func HandleStartFirstGame(bot *tgbotapi.BotAPI, chatID int64, roomId uint64) {
    var room structs.SquidRooms
    if err := engine.DB.Preload("Members.User").First(&room, roomId).Error; err != nil {
        return
    }

    room.Status = "game_red_green"
    room.LightStatus = "green" // Начинаем с зеленого света
    room.TimerSeconds = 180    // 3 минуты (3:00)
    engine.DB.Save(&room)

    // Формируем начальный текст игры с правилами
    text := buildGameText(&room, "🟢 ЗЕЛЕНЫЙ СВЕТ — Нужно идти (можно нажимать ШАГ)!")

    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "Markdown"

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("👟 ШАГ", "squid_step"),
        ),
    )
    msg.ReplyMarkup = keyboard

    sentMsg, err := bot.Send(msg)
    if err != nil {
        return
    }

    // Запускаем горутину для таймера и смены сигналов светофора в реальном времени
    go runGameLoop(bot, chatID, sentMsg.MessageID, roomId)
}

// runGameLoop управляет таймером (3:00) и случайным переключением света (ЗЕЛЕНЫЙ / КРАСНЫЙ)
func runGameLoop(bot *tgbotapi.BotAPI, chatID int64, messageID int, roomId uint64) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    lightSwitchCounter := 0

    for range ticker.C {
        var room structs.SquidRooms
        if err := engine.DB.Preload("Members.User").First(&room, roomId).Error; err != nil {
            return
        }

        if room.Status != "game_red_green" {
            return
        }

        if room.TimerSeconds > 0 {
            room.TimerSeconds--
        }

        lightSwitchCounter++
        if lightSwitchCounter >= 8 {
            lightSwitchCounter = 0
            if room.LightStatus == "green" {
                room.LightStatus = "red"
            } else {
                room.LightStatus = "green"
            }
        }
        engine.DB.Save(&room)

        if room.TimerSeconds <= 0 {
            endGameByTimeout(bot, chatID, messageID, roomId)
            return
        }

        var lightInstruction string
        if room.LightStatus == "green" {
            lightInstruction = "🟢 ЗЕЛЕНЫЙ СВЕТ — Нужно идти (можно нажимать ШАГ)!"
        } else {
            lightInstruction = "🔴 КРАСНЫЙ СВЕТ — СТОП! Не нажимайте ШАГ!"
        }

        updateGameMessageWithTimer(bot, chatID, messageID, &room, lightInstruction)
    }
}

// endGameByTimeout обрабатывает окончание времени (поражение всех оставшихся)
func endGameByTimeout(bot *tgbotapi.BotAPI, chatID int64, messageID int, roomId uint64) {
    var room structs.SquidRooms
    if err := engine.DB.Preload("Members.User").First(&room, roomId).Error; err != nil {
        return
    }

    engine.DB.Where("room_id = ?", roomId).Delete(&structs.SquidMembers{})

    room.Status = "closed"
    engine.DB.Save(&room)

    text := "⏰ *Время вышло (0:00)!*\n\n" +
        "К сожалению, таймер истек. Все оставшиеся участники проиграли в игре *Красный свет, зеленый свет*! ❌"

    editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
    editMsg.ParseMode = "Markdown"
    bot.Send(editMsg)
}

// buildGameText генерирует текст сообщения игры, включая правила, участников, прогресс-бар и таймер
func buildGameText(room *structs.SquidRooms, lightInstruction string) string {
    minutes := room.TimerSeconds / 60
    seconds := room.TimerSeconds % 60
    timerStr := fmt.Sprintf("%d:%02d", minutes, seconds)

    var participantsText string
    totalMembers := len(room.Members)

    if totalMembers > 0 {
        participantsText = "*Список участников:*\n"
        for i, member := range room.Members {
            name := member.User.FirstName
            if name == "" {
                name = fmt.Sprintf("ID: %d", member.UserId)
            }
            participantsText += fmt.Sprintf("%d. %s\n", i+1, name)
        }
    } else {
        participantsText = "_Все участники выбыли._\n"
    }

    progressPercent := (180 - room.TimerSeconds) * 100 / 180
    filledBlocks := progressPercent / 10
    if filledBlocks > 10 {
        filledBlocks = 10
    }
    emptyBlocks := 10 - filledBlocks
    progressBar := fmt.Sprintf("[%s%s] %d%%", repeatString("█", filledBlocks), repeatString("░", emptyBlocks), progressPercent)

    return fmt.Sprintf("🔴 *Первая игра: Красный свет, зеленый свет*\n\n"+
        "📜 *Правила игры:*\n"+
        "• %s\n"+
        "• Следите за таймером. Если время выйдет до финиша — вы проиграете.\n\n"+
        "⏱ *Оставшееся время:* `%s`\n\n"+
        "%s\n\n"+
        "📊 *Прогресс игры:* %s",
        lightInstruction, timerStr, participantsText, progressBar)
}

// updateGameMessageWithTimer обновляет интерфейс сообщения с учетом таймера и света
func updateGameMessageWithTimer(bot *tgbotapi.BotAPI, chatID int64, messageID int, room *structs.SquidRooms, lightInstruction string) {
    engine.DB.Preload("Members.User").First(room, room.ID)

    text := buildGameText(room, lightInstruction)

    editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
    editMsg.ParseMode = "Markdown"

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("👟 ШАГ", "squid_step"),
        ),
    )
    editMsg.ReplyMarkup = &keyboard

    bot.Send(editMsg)
}

// HandleSquidStepCallback обрабатывает нажатие кнопки "ШАГ"
func HandleSquidStepCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
    user_id := callback.From.ID

    var member structs.SquidMembers
    if err := engine.DB.Where("user_id = ?", uint64(user_id)).First(&member).Error; err != nil {
        answerCallback(bot, callback.ID, "❌ Вы не участвуете в игре.")
        return
    }

    var room structs.SquidRooms
    if err := engine.DB.Preload("Members.User").First(&room, member.RoomId).Error; err != nil {
        answerCallback(bot, callback.ID, "❌ Комната не найдена.")
        return
    }

    if room.LightStatus == "red" {
        if err := engine.DB.Delete(&member).Error; err != nil {
            log.Printf("DEBUG: [HandleSquidStepCallback] Ошибка удаления игрока: %v", err)
            answerCallback(bot, callback.ID, "❌ Ошибка выбывания из игры.")
            return
        }

        answerCallback(bot, callback.ID, "❌ Красный свет! Вы сделали шаг и выбыли из игры!")
        return
    }

    if room.LightStatus == "green" {
        answerCallback(bot, callback.ID, "✅ Успешно! Вы сделали шаг на зеленый свет.")
        return
    }

    answerCallback(bot, callback.ID, "⏳ Игра еще не переключила сигнал светофора.")
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

    case "squid_step":
        HandleSquidStepCallback(bot, callback)

    default:
        answerCallback(bot, callback.ID, "❓ Неизвестное действие.")
    }
}


func repeatString(s string, count int) string {
    result := ""
    for i := 0; i < count; i++ {
        result += s
    }
    return result
}

func HandleCloseRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_id := message.Chat.ID
    parsed_chat_id := std_helpers.ParseChatID(uint64(chat_id))
    user_id := message.From.ID

    // 1. Проверка прав пользователя в чате
    memberRole, err := helpers.GetMemberRole(uint64(user_id), uint64(parsed_chat_id))
    if err != nil {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет прав для выполнения этой команды."))
        return
    }

    var room structs.SquidRooms
    var queryErr error

    // Если пользователь — создатель или админ чата, он может закрыть любую открытую комнату.
    // Если обычный пользователь — только свою собственную (где он владелец).
    if std_helpers.IsUserOwnerOrCreator(&memberRole) {
        queryErr = engine.DB.Where("status = ?", "open").Order("id DESC").First(&room).Error
    } else {
        queryErr = engine.DB.Where("owner_id = ? AND status = ?", uint64(user_id), "open").First(&room).Error
    }

    if queryErr != nil {
        if errors.Is(queryErr, gorm.ErrRecordNotFound) {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ У вас нет активной открытой комнаты для закрытия."))
            return
        }
        log.Printf("DEBUG: [HandleCloseRoom] Ошибка поиска комнаты: %v", queryErr)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось найти комнату."))
        return
    }

    // 2. Меняем статус комнаты на закрытый
    room.Status = "closed"
    if err := engine.DB.Save(&room).Error; err != nil {
        log.Printf("DEBUG: [HandleCloseRoom] Ошибка обновления статуса комнаты: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось закрыть комнату."))
        return
    }

    msg := tgbotapi.NewMessage(chat_id, fmt.Sprintf("🔒 Комната `Игра в кальмара` (ID: `%d`) успешно закрыта!", room.ID))
    msg.ParseMode = "Markdown"
    bot.Send(msg)
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