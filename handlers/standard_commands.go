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
		"На ваш счет зачислено: `**%d монет**` 💰\n\n"+
		"📋 **Доступные команды:**\n"+
		"🔹 `/balance` — Узнать баланс монет\n"+
		"🔹 `/profile` — Просмотреть личный профиль\n"+
		"🔹 `/casino` `[ставка]` — Испытать удачу в слотах\n"+
		"🔹 `/roulette` `[ставка] [цвет]` — Сыграть в рулетку\n"+
		"🔹 `/bones` `[ставка]` — Бросить кости против бота 🎲\n\n"+
		"НОВОЕ:\n\n"+

		" `/promo active` `[code]` — Ввести промокод 🎲\n\n"+
		" `/promo create` `[code] [amount]` — Создать промокод с количеством монет 🎲\n\n"+
		" `/promo delete` `[code]` — Удалить промокод 🎲\n\n"+

		"Желаем удачи в игре! Пусть фортуна будет на вашей стороне! 🍀\n\n"+
		"Версия бота: 2.3 2026.05.28",
		message.From.FirstName, balance)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
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
