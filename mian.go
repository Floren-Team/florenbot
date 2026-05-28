package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"florenbot/bones"
	"florenbot/engine"
	"florenbot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// 1. Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Файл .env не найден, используются системные переменные")
	}

	// 2. Инициализация инфраструктуры
	engine.InitDB()
	engine.InitCache()

	// 3. Настройка бота
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ Переменная BOT_TOKEN не установлена")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации бота: %v", err)
	}

	bot.Debug = false
	log.Printf("🤖 Авторизовано как @%s", bot.Self.UserName)

	// 4. Канал для перехвата сигналов завершения (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 5. Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	startTime := time.Now()

	// 6. Запуск обработки событий в горутине
	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			// Защита от "старых" сообщений
			if update.Message.Time().Before(startTime) {
				continue
			}

			// Обработка новых участников
			if update.Message.NewChatMembers != nil {
				handleNewMembers(bot, update)
				continue
			}

			// Обработка команд: теперь проверяем оба префикса
			text := update.Message.Text
			if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "!") {
				handleCommands(bot, update.Message)
			}
		}
	}()

	// Блокируемся до получения сигнала выхода
	<-quit
	log.Println("🛑 Получен сигнал завершения. Останавливаю бота...")

	// Безопасное закрытие базы данных
	engine.CloseDB()
	engine.ShutdownCache()
	log.Println("✅ Бот успешно выключен.")
}

func handleNewMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	for _, newUser := range update.Message.NewChatMembers {
		isBanned, err := engine.IsUserBanned(newUser.ID)
		if err != nil {
			log.Printf("⚠️ Ошибка проверки ЧС: %v", err)
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

func handleCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	text := message.Text
	log.Printf("Сообщение от @%s: %s", message.From.UserName, text)

	// Проверяем, начинается ли сообщение с допустимого префикса
	// if !strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "!") {
	// 	log.Print("Неизвестный префикс команды")
	// 	return
	// }

	allowedPrefixes := []string{"/", "!"}

	foundPrefix := ""
	for _, p := range allowedPrefixes {
		if strings.HasPrefix(text, p) {
			foundPrefix = p
			break
		}
	}

	if foundPrefix == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Не могу обработать команду: используйте префиксы / или !"))
		return
	}

	

	// Разбиваем строку на части по пробелам
	parts := strings.Fields(text)

	// Получаем саму команду, убираем префикс (1 символ) и приводим к нижнему регистру
	command := strings.ToLower(parts[0][1:])

	switch command {
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
	case "promo":
		handlers.HandlePromo(bot, message)
	case "clan":
		handlers.HandleClan(bot, message)
	case "info":
		handlers.HandleInfo(bot, message)
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестная команда"))
	}
}