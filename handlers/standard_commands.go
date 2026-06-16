package handlers

import (
	"fmt"
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	consts "florenbot/consts"
	helpers "florenbot/engine/helpers"
)

// HandleStart обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	userProfile, err := helpers.GetUserByID(user_id)
	debug_type := GetEnvBool("DEBUG", false)
	if debug_type {
		log.Printf("Пользователь %s начал игровой клуб", message.From.FirstName)
		log.Printf("Результат: %v, %v", userProfile, err)
	}
	if err != nil {
		if debug_type {
			log.Printf("Ошибка при старте: %v", err)
		}
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить профиль"))
		return
	}

	if userProfile.Id == 0 {
		if debug_type {
			log.Printf("Пользователь %s создается...", message.From.FirstName)
		}
		userProfile, err = helpers.CreateUser(
			user_id,
			message.From.UserName,
			message.From.FirstName,
		)
		if err != nil {
			if debug_type {
				log.Printf("Ошибка при создании пользователя: %v", err)
				return
			}
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось создать пользователя"))
			return
		}

		if debug_type {
			log.Printf("Пользователь %s создан", message.From.FirstName)
			log.Printf("Результат: %v", userProfile)
		}
	} else {
		if debug_type {
			log.Printf("Пользователь %s уже существует", message.From.FirstName)
		}
	}

	// Красивое, структурированное приветствие с использованием эмодзи-разделителей
	text := fmt.Sprintf("👋 **Приветствуем вас, %s!**\n\n"+
		"Добро пожаловать в игровой клуб! 🎰\n"+
		"Текущий баланс: **`%.2f` $** 💰\n\n"+
		"📋 **Доступные команды:**\n"+
		"🔹 `/balance` — Узнать баланс\n"+
		"🔹 `/profile` — Просмотреть профиль\n"+
		"🔹 `/casino [ставка]` — Слоты\n"+
		"🔹 `/roulette [ставка] [цвет]` — Рулетка\n"+
		"🔹 `/bones [ставка]` — Кости 🎲\n\n"+
		"🔹 `/promo active [code]` — Активировать код\n"+
		"🔹 `/promo create [code] [amount]` — Создать код\n"+
		"🔹 `/promo delete [code]` — Удалить код\n\n"+
		"🔹 `/clan` — Управление кланами\n\n"+
		"Новые команды:\n\n"+
		"🔹 `/report` — Сделать отчет\n\n"+
		"Желаем удачи! 🍀\n\n"+
		"Версия бота: %s\nДата: %s\nВладелец: %s",
		message.From.FirstName, userProfile.Balance, consts.VERSION, consts.DATE_LAST_UPDATE, consts.OWNER)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func HandleInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Информация о боте\n\n"+
		"Версия: %s\n"+
		"Дата обновления: %s\n"+
		"Владелец: %s\n"+
		"GitHub: https://github.com/Floren-Team/florenbot", consts.VERSION, consts.DATE_LAST_UPDATE, consts.OWNER)))
}

// HandleBalance обрабатывает команду /balance
func HandleBalance(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	userProfile, err := helpers.GetUserByID(user_id)
	if err != nil {
		log.Printf("Ошибка получения баланса: %v", err)
		return
	}

	text := fmt.Sprintf("💰 **Ваш текущий баланс:** `%.2f` $.\n"+
		"Играйте в играх /casino, /roulette, /bones. И зарабатывайте деньги.", userProfile.Balance)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// HandleProfile обрабатывает команду /profile
func HandleProfile(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	user_id := uint64(message.From.ID)
	userProfile, err := helpers.GetUserByID(user_id)
	if err != nil {
		log.Printf("Ошибка профиля: %v", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить профиль"))
		return
	}

	negative_reputation := userProfile.NegativeReputation
	positive_reputation := userProfile.PositiveReputation

	total_reputation := negative_reputation + positive_reputation
	status_text := ""
	if positive_reputation >= 10 {
		userProfile.Status = 3
	} else if positive_reputation >= 5 {
		userProfile.Status = 2
	} else if negative_reputation >= 100 {
		userProfile.Status = 1
	} else {
		userProfile.Status = 0
	}

	switch userProfile.Status {
	case 3:
		status_text = "Отлично"
	case 2:
		status_text = "Хорошо"
	case 1:
		status_text = "Плохо"
	default:
		status_text = "Неизвестно"
	}

	switch userProfile.Role {
	case "admin":
		userProfile.Role = "Администратор"
	case "moderator":
		userProfile.Role = "Модератор"
	default:
		userProfile.Role = "Пользователь"
	}

	text := fmt.Sprintf("👤 **Это `%s`:**\n"+
		"└── **Баланс:** `%.2f` рублей 🪙\n"+
		"└── **Floren Coin:** `%.2f` монет 🪙\n"+
		"└── **Негативных репутации:** `%d`\n"+
		"└── **Роль:** `%s`\n"+
		"└── **Позитивных репутации:** `%d`\n\n"+
		"└── **Всего репутации:** `%d`\n"+
		"└── **Статус:** %s\n"+

		"Приятной вам игры в FlorenBot!",
		message.From.FirstName,
		userProfile.Balance,
		userProfile.FlorenCoin,
		userProfile.NegativeReputation,
		userProfile.Role,
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
