package handlers

import (
	consts "florenbot/consts"
	std_helpers "florenbot/helpers"
	helpers "florenbot/engine/helpers"
	engine "florenbot/engine/mysql"
	structs "florenbot/engine/structs"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleHelp(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    parsed_chat_id := std_helpers.ParseChatID(uint64(message.Chat.ID))
    user_id := uint64(message.From.ID)

    memberRole, err := helpers.GetMemberRole(user_id, uint64(parsed_chat_id))
    
    roleName := "Гость"
    if err == nil {
        roleName = memberRole.Name
    }

    info := fmt.Sprintf("👋 Привет! Твоя роль: *%s*\n\n", roleName)
    info += "📜 *Общие команды:*\n"
    info += "/help — Список команд\n"
    info += "/top — Топ сообщений\n"

    // 1. Команды для Модератора
    if memberRole.ShortName == "moderator" || std_helpers.IsUserAdmin(&memberRole) || std_helpers.IsUserOwnerOrCreator(&memberRole) {
        info += "\n🛡 *Панель модератора:*\n"
        info += "/ban — Ограничить пользователя\n"
        info += "/mute — Запретить общение\n"
        info += "/unmute — Снять мут\n"
    }

    // 2. Команды для Администратора (управление ролями)
    if std_helpers.IsUserAdmin(&memberRole) || std_helpers.IsUserOwnerOrCreator(&memberRole) {
        info += "\n👮 *Панель администратора:*\n"
    }

    // 3. Команды для Владельца и Создателя
    if std_helpers.IsUserOwnerOrCreator(&memberRole) {
        info += "\n👑 *Панель владельца:*\n"
        info += "/newrole [имя] [короткое] [служебное] — Создать роль\n"
        info += "/editrole [ID] [имя] [короткое] — Редактировать роль\n"
        info += "/delrole [ID] — Удалить роль\n"
        info += "/editcmd [ID] — Изменить доступ к команде\n"
    }

    // 4. Команды только для Создателя
    if std_helpers.IsUserCreator(&memberRole) {
        info += "\n⚙️ *Системные команды (Creator):*\n"
        info += "/sysrole — выдать системную роль\n"
    }

    msg := tgbotapi.NewMessage(message.Chat.ID, info)
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}

func HandleProfilePrivate(user_id uint64, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	userProfile, err := helpers.GetOrCreateUser(user_id, message.From.UserName, message.From.FirstName)
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
	chat_id := message.Chat.ID
    switch chat_type {
    case "private":
        // Вызываем функцию для обработки приватного профиля пользователя
        HandleProfilePrivate(uint64(message.From.ID), bot, message)

    case "group", "supergroup":
        // 1. Получаем пользователя из БД или создаем его, если он новый
        user, err := helpers.GetOrCreateUser(uint64(message.From.ID), message.From.UserName, message.From.FirstName)
        if err != nil {
            log.Printf("Ошибка при получении/создании пользователя: %v", err)
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось инициализировать ваш профиль."))
            return
        }

        // 2. Подготовка ID чата (преобразуем отрицательные ID Telegram в положительные для БД)
        chat_id := message.Chat.ID
        db_id := uint64(chat_id)
        if chat_id < 0 {
            db_id = uint64(-chat_id)
        }

        // 3. Создаем структуру чата для регистрации
        newChat := structs.Chat{
            ID:     db_id,
            Name:   message.Chat.Title,
            UserID: user.ID, // Привязываем чат к создателю
        }

        // 4. Регистрируем чат в базе данных
        err = helpers.CreateChat(newChat)
        if err != nil {
            log.Printf("Ошибка создания чата: %v", err)
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Не удалось создать запись чата."))
            return
        }

        // // 5. Инициализируем стандартные роли для этого чата
        err = helpers.InitDefaultRoles(db_id)
        if err != nil {
            log.Printf("Ошибка инициализации ролей: %v", err)
            bot.Send(tgbotapi.NewMessage(chat_id, "❌ Чат зарегистрирован, но произошла ошибка при настройке ролей."))
            return
        }

		

		err = helpers.AddMemberRole(user.ID, "owner", db_id)

        bot.Send(tgbotapi.NewMessage(chat_id, "✅ Чат успешно зарегистрирован!\n"+
                                            "Роли инициализированы! Просмотреть роли можно командой /roles"))

    default:
        // Если тип чата не поддерживается
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
    chatID := uint64(message.Chat.ID)
    parsed_chat_id := std_helpers.ParseChatID(uint64(chatID))


    // 1. Получаем профиль пользователя
    userProfile, err := helpers.GetUserByID(userID)
    if err != nil {
        log.Printf("Ошибка профиля: %v", err)
        bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось получить профиль"))
        return
    }

    // 2. Получаем роль пользователя в текущем чате
    var member structs.Member
    // Загружаем связь с ролью, чтобы получить её название
    err = engine.DB.Preload("Role").
        Where("user_id = ? AND chat_id = ?", userID, parsed_chat_id).
        First(&member).Error

    roleName := "Нет роли" // Значение по умолчанию
    if err == nil && member.Role.Name != "" {
        roleName = member.Role.Name
    }

    // 3. Расчет статуса
    newStatus := 0
    if userProfile.PositiveReputation >= 10 {
        newStatus = 3
    } else if userProfile.PositiveReputation >= 5 {
        newStatus = 2
    } else if userProfile.NegativeReputation >= 100 {
        newStatus = 1
    }

    // Обновляем статус в БД, если он изменился
    if userProfile.Status != newStatus {
        userProfile.Status = newStatus
        engine.DB.Model(&userProfile).Update("status", newStatus)
    }

    // 4. Определение текста статуса
    statusTexts := map[int]string{
        3: "Отлично",
        2: "Хорошо",
        1: "Плохо",
        0: "Неизвестно",
    }
    statusText := statusTexts[userProfile.Status]

    // 5. Расчет статистики
    totalReputation := userProfile.NegativeReputation + userProfile.PositiveReputation
    totalGames := userProfile.Wins + userProfile.Losses

    // 6. Формирование и отправка сообщения
    text := fmt.Sprintf("👤 **Профиль `%s`:**\n"+
        "└── **Баланс:** `%.2f` $\n"+
        "└── **Евро:** `%.2f` €\n"+
        "└── **Floren Coin:** `%.2f` монет\n"+
        "└── **Репутация:** `%d` (поз: %d, нег: %d)\n"+
        "└── **Роль:** `%s`\n"+
        "└── **Игры:** %d (Побед: %d, Поражений: %d)\n"+
        "└── **Статус:** %s\n\n"+
        "Приятной вам игры в FlorenBot!",
        userProfile.FirstName,
        userProfile.Balance,
        userProfile.Euro,
        userProfile.FlorenCoin,
        totalReputation,
        userProfile.PositiveReputation,
        userProfile.NegativeReputation,
        roleName,
        totalGames,
        userProfile.Wins,
        userProfile.Losses,
        statusText,
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
