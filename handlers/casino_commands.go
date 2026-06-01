package handlers

import (
	"florenbot/engine"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var StartRandomValue int

func init() {
	StartRandomValue = rand.Intn(100)
}

func GetEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// HandleCasino - Слоты (/casino 100)
func HandleCasino(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := message.CommandArguments()
	bet, err := strconv.Atoi(strings.TrimSpace(args))
	debug_type := GetEnvBool("DEBUG", false)
	if err != nil || bet <= 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректную сумму ставки числом. Пример: `/casino 100`")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	user_id := uint64(message.From.ID)

	balance, err := engine.GetBalance(user_id, message.From.UserName)
	if err != nil {
		if debug_type {
			log.Printf("Ошибка получения баланса: %v", err)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось проверить ваш баланс.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if bet < 50 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальная ставка - 50 монет")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Не удалось отправить сообщение: %v", err)
		}
		return
	}

	if bet > balance {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ У вас недостаточно монет! Ваш баланс: %d", balance))); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	elements := []string{"🍒", "🍋", "🎰", "💎"}
	res1 := elements[rand.Intn(len(elements))]
	res2 := elements[rand.Intn(len(elements))]
	res3 := elements[rand.Intn(len(elements))]

	resultStr := fmt.Sprintf("🎰 Крутим барабаны:\n[ %s | %s | %s ]\n\n", res1, res2, res3)

	if res1 == res2 && res2 == res3 {
		winAmount := bet * 5
		engine.ChangeBalance(user_id, winAmount)
		resultStr += fmt.Sprintf("🎉 ДЖЕКПОТ! Вы выиграли %d монет!", winAmount)
	} else if res1 == res2 || res2 == res3 || res1 == res3 {
		winAmount := bet * 1 // Возврат ставки + выигрыш x2 чистыми
		engine.ChangeBalance(user_id, winAmount)
		resultStr += fmt.Sprintf("💵 Победа! Вы выиграли %d монет!", bet*2)
	} else {
		engine.ChangeBalance(user_id, -bet)
		resultStr += fmt.Sprintf("📉 Проигрыш. Вы потеряли %d монет.", bet)
	}

	if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, resultStr)); err != nil {
		log.Printf("Не удалось отправить сообщение: %v", err)
	}
}

// HandleRoulette - Рулетка (/roulette 100 красное)
func HandleRoulette(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	user_id := uint64(message.From.ID)
	// Если аргументов меньше 2 (например, ввели только ставку или вообще буквы)
	if len(args) < 2 {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите корректную ставку и цвет. Пример: `/roulette 100 красное`")); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	bet, err := strconv.Atoi(args[0])
	if err != nil || bet <= 0 {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ставка должна быть положительным числом! Пример: `/roulette 50 черное`")); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	choice := strings.ToLower(args[1])
	if choice != "красное" && choice != "черное" && choice != "зеленое" {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверно указан цвет. Выберите один из трех: `красное`, `черное` или `зеленое`")); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	balance, err := engine.GetBalance(user_id, message.From.UserName)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка авторизации игрового профиля.")
		msg.ReplyToMessageID = message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if bet < 50 {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Минимальная ставка - 50 монет")); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
		return
	}

	if bet > balance {
		if _, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Недостаточно монет. Ваш баланс: %d", balance))); err != nil {
			log.Printf("Ошибка отправки уведомления: %v", err)
		}
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
		engine.ChangeBalance(user_id, winAmount)
		output += fmt.Sprintf("🎉 Вы угадали! Выигрыш: +%d монет.", winAmount+bet)
	} else {
		engine.ChangeBalance(user_id, -bet)
		output += fmt.Sprintf("📉 Не повезло. Ставка потеряна: -%d монет.", bet)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, output)
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки уведомления: %v", err)
	}
}
