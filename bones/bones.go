package bones

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"florenbot/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBones обрабатывает команду /bones [ставка]
func HandleBones(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	bet, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || bet <= 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректную ставку. Пример: `/bones 100`"))
		return
	}

	user_id := uint64(message.From.ID)

	// Проверяем баланс игрока
	balance, err := engine.GetBalance(user_id, message.From.UserName)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить ваш баланс."))
		return
	}

	if bet > balance {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Недостаточно монет! Ваш баланс: %d", balance)))
		return
	}

	if bet < 30 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальное число ставки - 30"))
		return
	}

	// 1. Бросок игрока
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("%s бросает кости...", message.From.FirstName)))
	playerDiceMsg := tgbotapi.NewDice(message.Chat.ID) // Отправляет анимированный кубик
	playerResult, err := bot.Send(playerDiceMsg)
	if err != nil {
		return
	}
	playerValue := playerResult.Dice.Value // Число от 1 до 6

	// Небольшая задержка для реалистичности, пока крутится кубик игрока
	time.Sleep(3 * time.Second)

	// 2. Бросок бота
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Бот бросает кости..."))
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
		engine.ChangeBalance(user_id, bet)
		resultText += fmt.Sprintf(" **Победа!** Вы оказались удачливее бота и выиграли **%d монет**!", bet)
	} else if playerValue < botValue {
		// Игрок проиграл
		engine.ChangeBalance(user_id, -bet)
		resultText += fmt.Sprintf(" **Проигрыш.** Бот победил. Вы потеряли **%d монет**.", bet)
	} else {
		// Ничья (деньги не списываются)
		resultText += "🤝 **Ничья!** Силы равны, монеты остаются при вас."
	}

	// Отправляем финальный вердикт
	msg := tgbotapi.NewMessage(message.Chat.ID, resultText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
