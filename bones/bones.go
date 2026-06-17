package bones

import (
	"fmt"
	"log"
	"strconv"
	"time"

	helpers "florenbot/engine/helpers"
	engine "florenbot/engine/cache"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBones обрабатывает команду /bones [ставка]
func HandleBones(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	bet, err := strconv.ParseFloat(args, 64)
	bet = float64(bet)
	msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректную ставку. Пример: `/bones 100`")
	msg.ReplyToMessageID = message.MessageID
	if err != nil || bet <= 0 {
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	user_id := uint64(message.From.ID)

	// Проверяем баланс игрока
	balance, err := engine.GetBalance(user_id, message.From.UserName)
	msg = tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить ваш баланс.")
	if err != nil {
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if bet > balance {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ У вас недостаточно монет! Ваш баланс: %.2f", balance))
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		// bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Недостаточно монет! Ваш баланс: %d", balance)))
		return
	}

	if bet < 30 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальное число ставки - 30")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		// bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальное число ставки - 30"))
		return
	}

	// 1. Бросок игрока
	msg = tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("%s бросает кости...", message.From.FirstName))
	msg.ReplyToMessageID = message.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки уведомления: %v", err)
	}
	playerDiceMsg := tgbotapi.NewDice(message.Chat.ID) // Отправляет анимированный кубик
	playerResult, err := bot.Send(playerDiceMsg)
	if err != nil {
		return
	}
	playerValue := playerResult.Dice.Value // Число от 1 до 6

	// Небольшая задержка для реалистичности, пока крутится кубик игрока
	time.Sleep(3 * time.Second)

	// 2. Бросок бота
	msg = tgbotapi.NewMessage(message.Chat.ID, "🤖 Бот бросает кости...")
	msg.ReplyToMessageID = message.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки уведомления: %v", err)
	}

	botDiceMsg := tgbotapi.NewDice(message.Chat.ID)
	botResult, err := bot.Send(botDiceMsg)
	if err != nil {
		return
	}
	botValue := botResult.Dice.Value // Число от 1 до 6

	// Снова ждем анимацию кубика бота
	time.Sleep(3 * time.Second)

	// 3. Сравнение результатов
	resultText := fmt.Sprintf("📊 **Итоги раунда:**\nВы выбросили: %d\nБот выбросил: %d\n\n", playerValue, botValue)

	if playerValue > botValue {
		// Игрок выиграл (получает x2 от ставки, то есть чистый плюс равен ставке)
		if err := engine.ChangeBalance(user_id, bet); err != nil {
			log.Printf("Ошибка изменения баланса: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось изменить ваш баланс."))
			return
		}
		err := helpers.AddUserToLosses(user_id)
		if err != nil {
			log.Printf("Ошибка изменения баланса: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка изменения рекордов :("))
			return
		}
		resultText += fmt.Sprintf(" **Победа!** Вы оказались удачливее бота и выиграли **%.2f монет**!", bet)
	} else if playerValue < botValue {
		// Игрок проиграл
		if err := engine.ChangeBalance(user_id, -bet); err != nil {
			log.Printf("Ошибка изменения баланса: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось изменить ваш баланс."))
			return
		}
		err := helpers.AddUserToLosses(user_id)
		if err != nil {
			log.Printf("Ошибка изменения баланса: %v", err)
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка изменения рекордов :("))
			return
		}
		resultText += fmt.Sprintf(" **Проигрыш.** Бот победил. Вы потеряли **%2.f монет**.", bet)
	} else {
		// Ничья (деньги не списываются)
		resultText += "🤝 **Ничья!** Силы равны, монеты остаются при вас."
	}

	// Отправляем финальный вердикт
	msg = tgbotapi.NewMessage(message.Chat.ID, resultText)
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки уведомления: %v", err)
	}
}
