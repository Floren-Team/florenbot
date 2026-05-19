package main

import (
	"fmt"
	"log"
	"os"

	"florenbot/bones"
	"florenbot/engine"
	"florenbot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// 1. ПРОВЕРКА ПРИ ВСТУПЛЕНИИ В ЧАТ
		if update.Message.NewChatMembers != nil {
			for _, newUser := range update.Message.NewChatMembers { // Исправлено: убрана лишняя {
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
			continue
		}

		// 2. ОБРАБОТКА КОМАНД
		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "start":
			handlers.HandleStart(bot, update.Message)
		case "balance":
			handlers.HandleBalance(bot, update.Message)
		case "profile":
			handlers.HandleProfile(bot, update.Message)
		case "casino":
			handlers.HandleCasino(bot, update.Message)
		case "roulette":
			handlers.HandleRoulette(bot, update.Message)
		case "bones":
			bones.HandleBones(bot, update.Message)
		case "kick":
			handlers.HandleKick(bot, update.Message)
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Неизвестная команда"))
		}
	}
}