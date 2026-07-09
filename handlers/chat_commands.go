package handlers

import (
    "fmt"
    "log"
    "strings"

    helpers "florenbot/engine/helpers"
    structs "florenbot/engine/structs"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleChat(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    args := message.CommandArguments()
    parts := strings.Fields(args)
    chat_id := message.Chat.ID
    
    // Безопасное получение ID для БД (всегда положительное)
    db_id := chat_id
    if db_id < 0 {
        db_id = -db_id
    }

    if len(parts) < 1 {
        bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверный формат! Используйте: `/chat [create|delete|get]`"))
        return
    }

    action := parts[0]

    switch action {
    case "create":
        {
            user_id := uint64(message.From.ID)
            newChat := structs.Chat{
                ID:     uint(db_id), // Сохраняем положительный ID
                Name:   message.Chat.Title,
                UserID: int64(user_id),
            }

            err := helpers.CreateChat(newChat)
            if err != nil {
                log.Printf("Ошибка создания чата: %v", err)
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось создать запись чата"))
                return
            }

            bot.Send(tgbotapi.NewMessage(chat_id, "✅ Чат создан"))
        }
    case "delete":
        {
            chat, err := helpers.GetChatById(db_id)
            if err != nil {
                log.Printf("Ошибка получения чата: %v", err)
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Чат не найден"))
                return
            }

            if chat.UserID != int64(message.From.ID) {
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Вы не владелец чата."))
                return
            }

            err = helpers.DeleteChat(db_id)
            if err != nil {
                log.Printf("Ошибка удаления чата: %v", err)
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Ошибка при удалении"))
                return
            }
            bot.Send(tgbotapi.NewMessage(chat_id, "✅ Чат удален"))
        }
    case "get":
        {
            if message.Chat.Type == "private" {
                bot.Send(tgbotapi.NewMessage(chat_id, "Команда /chat get доступна только в группах"))
                return
            }
            
            chat, err := helpers.GetChatById(db_id)
            if err != nil {
                log.Printf("Ошибка получения чата: %v", err)
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Чат не найден в системе"))
                return
            }
            
            user, err := helpers.GetUser(uint64(chat.UserID))
            if err != nil {
                log.Printf("Ошибка получения владельца: %v", err)
                bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось получить владельца"))
                return
            }

            info := fmt.Sprintf("Информация о чате:\n\n"+
                               "Имя: %s\n"+
                               "Владелец: %s", chat.Name, user.FirstName)

            bot.Send(tgbotapi.NewMessage(chat_id, info))
        }
    default:
        {
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Неверная команда! Используйте: `/chat [create|delete|get]`"))
        }
    }
}