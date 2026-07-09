package handlers

import (
	consts "florenbot/consts"
	helpers "florenbot/engine/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	structs "florenbot/engine/structs"
	engine "florenbot/engine/mysql"
)



func HandleProfilePrivate(user_id uint64, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
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

	if userProfile.ID == 0 {
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

func HandleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    chat_type := message.Chat.Type

    switch chat_type {
    case "private":
        // Вызываем функцию для приватного чата
        HandleProfilePrivate(uint64(message.From.ID), bot, message)

    case "group", "supergroup":
        // Логика регистрации чата
        user_id := uint64(message.From.ID)
        chat_id := message.Chat.ID

        db_id := chat_id
        if db_id < 0 {
            db_id = -db_id
        }
        
        newChat := structs.Chat{
            ID:     uint64(db_id),
            Name:   message.Chat.Title,
            UserID: uint64(user_id),
        }

        err := helpers.CreateChat(newChat)
        if err != nil {
            log.Printf("Ошибка создания чата: %v", err)
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось создать запись чата"))
            return
        }

        bot.Send(tgbotapi.NewMessage(chat_id, "✅ Чат успешно зарегистрирован!"))

    default:
        // Если тип чата не определен
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестный тип чата."))
    }
}

func HandleInfo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Информация о боте\n\n"+
		"Версия: %s\n"+
		"Дата обновления: %s\n"+
		"Владелец: %s\n"+
		"GitHub: %s", consts.VERSION, consts.DATE_LAST_UPDATE, consts.OWNER, consts.REPO_URL)))
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
    userID := uint64(message.From.ID)
    userProfile, err := helpers.GetUserByID(userID)
    if err != nil {
        log.Printf("Ошибка профиля: %v", err)
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить профиль"))
        return
    }

    // 1. Расчет статуса
    newStatus := 0
    if userProfile.PositiveReputation >= 10 {
        newStatus = 3
    } else if userProfile.PositiveReputation >= 5 {
        newStatus = 2
    } else if userProfile.NegativeReputation >= 100 {
        newStatus = 1
    }

    // Обновляем в БД, если статус изменился
    if userProfile.Status != newStatus {
        userProfile.Status = newStatus
        engine.DB.Model(&userProfile).Update("status", newStatus)
    }

    // 2. Определение текста статуса
    statusTexts := map[int]string{
        3: "Отлично",
        2: "Хорошо",
        1: "Плохо",
        0: "Неизвестно",
    }
    statusText := statusTexts[userProfile.Status]
    roleName := "Пользователь"
    if userProfile.RoleID != nil && userProfile.Role.ID != 0 {
        roleName = userProfile.Role.Name 
    }

    // 4. Расчет статистики
    totalReputation := userProfile.NegativeReputation + userProfile.PositiveReputation
    totalGames := userProfile.Wins + userProfile.Losses

    // 5. Формирование сообщения
  text := fmt.Sprintf("👤 **Профиль `%s`:**\n"+
        "└── **Баланс:** `%.2f` $\n"+
        "└── **Евро:** `%.2f` €\n"+
        "└── **Floren Coin:** `%.2f` монет\n"+
        "└── **Репутация:** `%d` (поз: %d, нег: %d)\n"+
        "└── **Роль:** `%s`\n"+
        "└── **Игры:** %d (Побед: %d, Поражений: %d)\n"+
        "└── **Статус:** %s\n\n"+
        "Приятной вам игры в FlorenBot!",
        userProfile.FirstName,        // %s
        userProfile.Balance,          // %.2f
        userProfile.Euro,             // %.2f
        userProfile.FlorenCoin,       // %.2f
        totalReputation,              // %d
        userProfile.PositiveReputation, // %d
        userProfile.NegativeReputation, // %d
        roleName,                     // %s
        totalGames,                   // %d
        userProfile.Wins,             // %d
        userProfile.Losses,           // %d
        statusText,                   // %s
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

	// Пытаемся кикнуть пользователя
	_, err := bot.Request(kickConfig)
	if err != nil {
		log.Printf("Ошибка кика: %v", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось кикнуть пользователя. Возможно, у бота нет прав или пользователь админ."))
		return
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "До свидания!"))
}
