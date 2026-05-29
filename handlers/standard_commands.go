package handlers

import (
	"fmt"
	"log"

	"florenbot/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	balance, err := engine.GetBalance(user_id, message.From.UserName)
	if err != nil {
		log.Printf("Ошибка при старте: %v", err)
		return
	}

	// Красивое, структурированное приветствие с использованием эмодзи-разделителей
	text := fmt.Sprintf("👋 **Приветствуем вас, %s!**\n\n"+
		"Добро пожаловать в игровой клуб! 🎰\n"+
		"Текущий баланс: **%d рублей** 💰\n\n"+
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
		"Версия бота: 2.5\nДата: 2026.05.28",
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
		"Поддержка: Hamster Bot Владелец: @Serh1t"))
}

// HandleBalance обрабатывает команду /balance
func HandleBalance(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	balance, err := engine.GetBalance(user_id, message.From.UserName)
	if err != nil {
		log.Printf("Ошибка получения баланса: %v", err)
		return
	}

	text := fmt.Sprintf("💰 **Ваш текущий баланс:** `%d` рублей.", balance)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// HandleProfile обрабатывает команду /profile
func HandleProfile(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	userProfile, err := engine.GetUserByID(user_id)
	if err != nil {
		log.Printf("Ошибка профиля: %v", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить профиль"))
		return
	}

	negative_reputation := userProfile.NegativeReputation
	positive_reputation := userProfile.PositiveReputation

	total_reputation := negative_reputation + positive_reputation
	status_text := ""
	if total_reputation >= 10 {
		userProfile.Status = 3
	} else if total_reputation >= 5 {
		userProfile.Status = 2
	} else if total_reputation >= 1 {
		userProfile.Status = 1
	} else {
		userProfile.Status = 0
	}

	if userProfile.Status == 3 {
		status_text = "Отлично"
	} else if userProfile.Status == 2 {
		status_text = "Хорошо"
	} else if userProfile.Status == 1 {
		status_text = "Нормально"
	} else {
		status_text = "Неизвестно"
	}

	text := fmt.Sprintf("👤 **Это %s:**\n"+
		"└── **Баланс:** %.2f рублей 🪙\n"+
		"└── **Негативных репутации:** %d\n"+
		"└── **Позитивных репутации:** %d\n\n"+
		"└── **Всего репутации:** %d\n"+
		"└── **Статус:** %d\n"+

		"Приятной вам игры в FlorenBot!",
		message.From.FirstName,
		userProfile.Balance,
		userProfile.NegativeReputation,
		userProfile.PositiveReputation,
		total_reputation,
		status_text,
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
