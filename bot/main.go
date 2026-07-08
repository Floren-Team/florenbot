package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"florenbot/consts"
	helpers "florenbot/helpers"

	"flag"
	"florenbot/bones"
	cache "florenbot/engine/cache"
	helper "florenbot/engine/helpers"
	engine "florenbot/engine/mysql"
	"florenbot/handlers"
	admin_handlers "florenbot/handlers/admin"
	"runtime"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("env not found, using system environment variables")
	}
	verify := flag.Bool("skip", false, "Пропустить проверки хеширования")
	version := flag.Bool("version", false, "Версия бота")
	flag.Parse()

	if *version {
		log.Printf("Версия: %s", consts.VERSION)
		os.Exit(0)
	}

	if *verify {
		log.Println("⚠️ Пропускаем проверку хеширования (режим --skip)")
	} else {
		log.Println("Проверка хеширования...")
		if err := helpers.CheckHashAndGpg(); err != nil {
			panic(err)
		}
	}

	log.Printf("ОС: %s, Архитектура: %s", runtime.GOOS, runtime.GOARCH)

	time.Sleep(4 * time.Second)

	engine.ConnectDB()
	cache.InitCache()

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	startTime := time.Now()

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.Time().Before(startTime) {
				continue
			}

			if update.Message.NewChatMembers != nil {
				handleNewMembers(bot, update)
				continue
			}

			handleMessage(bot, update.Message)
		}
	}()

	<-quit
	log.Println("🛑 Останавливаю бота...")
	engine.CloseDB()
	cache.ShutdownCache()
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	text := message.Text
	lowerText := strings.ToLower(text)

	if strings.Contains(lowerText, "спасибо") {
		log.Printf("Зафиксирована благодарность от @%s", message.From.UserName)
		handlers.HandleThanks(bot, message)
		return
	}

	if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "!") {
		handleCommands(bot, message)
	}
}

func handleCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	parts := strings.Fields(message.Text)
	if len(parts) == 0 {
		return
	}

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
	case "clan":
		handlers.HandleClan(bot, message)
	case "roulette":

		handlers.HandleRoulette(bot, message)

	case "bones":

		bones.HandleBones(bot, message)

	case "q":

		handlers.HandleQuit(bot, message)

	case "newsletter":
		admin_handlers.HandleNewsLetter(bot, message)

	case "chat":
		handlers.HandleChat(bot, message)

	case "promo":

		handlers.HandlePromo(bot, message)

	case "info":

		handlers.HandleInfo(bot, message)

	case "rep":

		handlers.HandleReputation(bot, message)
	case "report":
		handlers.HandleReport(bot, message)
	case "спасибо":
		handlers.HandleThanks(bot, message)
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестная команда"))
	}
}

func handleNewMembers(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	for _, newUser := range update.Message.NewChatMembers {
		user_id := newUser.ID
		isBanned := helper.IsUserBanned(uint64(user_id))
		if isBanned {
			bot.Request(tgbotapi.BanChatMemberConfig{ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: update.Message.Chat.ID, UserID: newUser.ID}})
		}
	}
}
