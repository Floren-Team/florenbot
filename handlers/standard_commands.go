package handlers

import (
	"fmt"
	"log"

	"florenbot/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	balance, err := engine.GetBalance(message.From.ID, message.From.UserName)
	if err != nil {
		log.Printf("Ошибка при старте: %v", err)
		return
	}

	// Красивое, структурированное приветствие с использованием эмодзи-разделителей
	text := fmt.Sprintf("👋 **Приветствуем вас, %s!**\n\n"+
    "Добро пожаловать в игровой клуб! 🎰\n"+
    "На ваш счет зачислено: **%d монет** 💰\n\n"+
    "📋 **Доступные команды:**\n"+
    "🔹 `/balance` — Узнать баланс\n"+
    "🔹 `/profile` — Просмотреть профиль\n"+
    "🔹 `/casino [ставка]` — Слоты\n"+
    "🔹 `/roulette [ставка] [цвет]` — Рулетка\n"+
    "🔹 `/bones [ставка]` — Кости 🎲\n\n"+
    "🔹 `/promo active [code]` — Активировать код\n"+
    "🔹 `/promo create [code] [amount]` — Создать код\n"+
    "🔹 `/promo delete [code]` — Удалить код\n\n"+
    "🆕 **НОВОЕ:**\n"+
    "🔹 `/clan` — Управление кланами\n\n"+
    "Желаем удачи! 🍀\n\n"+
    "Версия бота: 2.4\nДата: 2026.05.28",
	message.From.FirstName, balance)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Информация о боте\n\n"+
		"Версия: 2.3\n"+
		"Дата обновления: 2026.05.28\n"+
		"Автор: Egor Luchiy\n"+
		"GitHub: -\n"+
		"Поддежрка Hamster Bot Владелец: @Serh1t"))
}

// HandleBalance обрабатывает команду /balance
func HandleBalance(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	balance, err := engine.GetBalance(message.From.ID, message.From.UserName)
	if err != nil {
		log.Printf("Ошибка получения баланса: %v", err)
		return
	}

	text := fmt.Sprintf("💰 **Ваш текущий баланс:** `%d` монет.", balance)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// HandleProfile обрабатывает команду /profile
func HandleProfile(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	balance, err := engine.GetBalance(message.From.ID, message.From.UserName)
	if err != nil {
		log.Printf("Ошибка профиля: %v", err)
		return
	}

	text := fmt.Sprintf("👤 **Ваш игровой профиль:**\n"+
		"├── **Имя:** %s\n"+
		"├── **ID:** `%d`\n"+
		"└── **Баланс:** %d монет 🪙\n"+
		"Приятной вам игры в FlorenBot!",
		message.From.FirstName,
		message.From.ID,
		balance,
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleQuit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {

	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: message.Chat.ID,
			UserID: message.From.ID,
		},
	}

	// Оголошуємо err перед використанням
	_, err := bot.Request(kickConfig)
	if err != nil {
		log.Printf("Ошибка кика: %v", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось кикнуть пользователя. Возможно, у бота нет прав или пользователь админ."))
		return
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "До свидания!"))
}
