package handlers

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleKick(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    // 1. Получаем список админов
    admins, err := bot.GetChatAdministrators(tgbotapi.ChatAdministratorsConfig{
        ChatConfig: tgbotapi.ChatConfig{ChatID: message.Chat.ID},
    })

    if err != nil {
        log.Printf("Ошибка получения админов: %v", err)
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка доступа к списку администраторов"))
        return // ВАЖНО: обязательно выходим, если произошла ошибка
    }

    // 2. Проверка прав отправителя
    isAdmin := false
    for _, admin := range admins {
        if admin.User.ID == message.From.ID {
            isAdmin = true
            break
        }
    }

    if !isAdmin {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⛔️ У вас нет прав администратора"))
        return
    }

    // 3. Проверка наличия ответа на сообщение
    if message.ReplyToMessage == nil {
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Нужно ответить на сообщение пользователя, которого хочешь кикнуть"))
        return
    }

    // 4. Кик пользователя (BanChatMember)
    // Telegram API теперь использует BanChatMemberConfig вместо KickChatMemberConfig
    kickConfig := tgbotapi.BanChatMemberConfig{
        ChatMemberConfig: tgbotapi.ChatMemberConfig{
            ChatID: message.Chat.ID,
            UserID: message.ReplyToMessage.From.ID,
        },
    }

    _, err = bot.Request(kickConfig)
    if err != nil {
        log.Printf("Ошибка кика: %v", err)
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось кикнуть пользователя. Возможно, у бота нет прав или пользователь админ."))
        return
    }

    bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ Пользователь исключен"))
}