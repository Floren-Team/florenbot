package handlers

import (
	helpers "florenbot/engine/helpers"
	std_helpers "florenbot/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	structs "florenbot/engine/structs"
	engine "florenbot/engine/mysql"
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

	exists, err := helpers.GetRoomSquid(engine.DB, uint64(user_id)) 
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
	user_id := message.From.ID

	exists, err := helpers.GetRoomSquid(engine.DB, uint64(user_id)) 
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
	_, err = helpers.GetMemberRole(uint64(message.From.ID), uint64(parsed_chat_id))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chat_id, "❌ Вы не можете посещать комнату."))
		return
	}
	msg := tgbotapi.NewMessage(chat_id, "Вы вошли в комнату `Игра в кальмара`.")
	msg.ParseMode = "Markdown"

	err = helpers.JoinRoom(engine.DB, uint64(user_id), uint64(parsed_chat_id))
    if err != nil {
        log.Printf("DEBUG: [HandleJoinRoom] Ошибка вошедшего в комнату: %v", err)
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось войти в комнату."))
        return
    }

	bot.Send(msg)



	
}