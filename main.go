package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"florenbot/bones"
	"florenbot/engine"
	"florenbot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}

	// Инициализация инфраструктуры
	engine.InitDB()
	defer engine.DB.Close()
	engine.InitRedis()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Переменная BOT_TOKEN не установлена в .env")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	bot.Debug = false
	log.Printf("Авторизовано под аккаунтом %s", bot.Self.UserName)

	// Настройка получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Время запуска бота, чтобы игнорировать сообщения, пришедшие до старта
	startTime := time.Now()

	for update := range updates {
		// Пропускаем, если сообщение пустое
		if update.Message == nil {
			continue
		}

		// ЗАЩИТА: Игнорируем сообщения, пришедшие до запуска бота
		// Это решает проблему "просыпания" бота и ответов на старые запросы
		if update.Message.Time().Before(startTime) {
			continue
		}

		// 1. ПРОВЕРКА ПРИ ВСТУПЛЕНИИ В ЧАТ
		if update.Message.NewChatMembers != nil {
			handleNewMembers(bot, update)
			continue
		}

		// 2. ОБРАБОТКА КОМАНД
		if update.Message.IsCommand() {
			handleCommands(bot, update.Message)
		}
	}
}

// handleNewMembers — отдельная функция для проверки ЧС
func handleNewMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	for _, newUser := range update.Message.NewChatMembers {
		isBanned, err := engine.IsUserBanned(newUser.ID)
		if err != nil {
			log.Printf("Ошибка проверки ЧС: %v", err)
			continue
		}

		if isBanned {
			kickConfig := tgbotapi.BanChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: update.Message.Chat.ID,
					UserID: newUser.ID,
				},
			}
			bot.Request(kickConfig)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("🚫 Пользователь @%s в черном списке. Исключен.", newUser.UserName)))
		}
	}
}

// handleCommands — централизованный обработчик команд
func handleCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		handlers.HandleStart(bot, message)
	case "balance":
		handlers.HandleBalance(bot, message)
	case "profile":
		handlers.HandleProfile(bot, message)
	case "casino":
		handlers.HandleCasino(bot, message)
	case "roulette":
		handlers.HandleRoulette(bot, message)
	case "bones":
		bones.HandleBones(bot, message)
	case "q":
		handlers.HandleQuit(bot, message)
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестная команда"))
	}
}