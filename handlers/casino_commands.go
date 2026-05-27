package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"florenbot/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// HandleCasino - Слоты (/casino 100)
func HandleCasino(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	bet, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || bet <= 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректную сумму ставки числом. Пример: `/casino 100`"))
		return
	}

	balance, err := engine.GetBalance(message.From.ID, message.From.UserName)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить ваш баланс."))
		return
	}

	if bet < 50 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальная ставка - 50 монет"))
		return
	}

	if bet > balance {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ У вас недостаточно монет! Ваш баланс: %d", balance)))
		return
	}

	elements := []string{"🍒", "🍋", "🎰", "💎"}
	res1 := elements[rand.Intn(len(elements))]
	res2 := elements[rand.Intn(len(elements))]
	res3 := elements[rand.Intn(len(elements))]

	resultStr := fmt.Sprintf("🎰 Крутим барабаны:\n[ %s | %s | %s ]\n\n", res1, res2, res3)

	if res1 == res2 && res2 == res3 {
		winAmount := bet * 5
		engine.ChangeBalance(message.From.ID, winAmount)
		resultStr += fmt.Sprintf("🎉 ДЖЕКПОТ! Вы выиграли %d монет!", winAmount)
	} else if res1 == res2 || res2 == res3 || res1 == res3 {
		winAmount := bet * 1 // Возврат ставки + выигрыш x2 чистыми
		engine.ChangeBalance(message.From.ID, winAmount)
		resultStr += fmt.Sprintf("💵 Победа! Вы выиграли %d монет!", bet*2)
	} else {
		engine.ChangeBalance(message.From.ID, -bet)
		resultStr += fmt.Sprintf("📉 Проигрыш. Вы потеряли %d монет.", bet)
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, resultStr))
}

// HandleRoulette - Рулетка (/roulette 100 красное)
func HandleRoulette(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	// Если аргументов меньше 2 (например, ввели только ставку или вообще буквы)
	if len(args) < 2 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат!\nИспользуйте: `/roulette [ставка] [красное/черное/зеленое]`\n\nПример: `/roulette 100 красное`"))
		return
	}

	bet, err := strconv.Atoi(args[0])
	if err != nil || bet <= 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ставка должна быть положительным числом! Пример: `/roulette 50 черное`"))
		return
	}

	choice := strings.ToLower(args[1])
	if choice != "красное" && choice != "черное" && choice != "зеленое" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверно указан цвет. Выберите один из трех: `красное`, `черное` или `зеленое`"))
		return
	}

	balance, err := engine.GetBalance(message.From.ID, message.From.UserName)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка авторизации игрового профиля."))
		return
	}

	if bet < 50 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальная ставка - 50 монет"))
		return
	}

	if bet > balance {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Недостаточно монет. Ваш баланс: %d", balance)))
		return
	}

	spin := rand.Intn(37)
	var landedColor string

	if spin == 0 {
		landedColor = "зеленое"
	} else if spin%2 == 0 {
		landedColor = "красное"
	} else {
		landedColor = "черное"
	}

	output := fmt.Sprintf("🔮 Шарик остановился на: **%s** (%d)\n\n", landedColor, spin)

	if choice == landedColor {
		multiplier := 1
		if landedColor == "зеленое" {
			multiplier = 35
		}
		winAmount := bet * multiplier
		engine.ChangeBalance(message.From.ID, winAmount)
		output += fmt.Sprintf("🎉 Вы угадали! Выигрыш: +%d монет.", winAmount+bet)
	} else {
		engine.ChangeBalance(message.From.ID, -bet)
		output += fmt.Sprintf("📉 Не повезло. Ставка потеряна: -%d монет.", bet)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, output)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
