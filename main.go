package main

import (
	"context"
	"github.com/dstotijn/go-notion"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"sync"
)

const offset = 0 //??

var (
	log  *slog.Logger
	once sync.Once
)

func main() {
	_ = GetLogger() // set log = logger, and create global logger

	err := godotenv.Load()
	if err != nil {
		log.Warn("fail to read from .env file", "err", err)
	}

	tgToken := os.Getenv("TELEGRAM_SECRET")
	if len(tgToken) == 0 {
		log.Error("telegram secret is not specified")
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Error("fail to create bot", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	notionSecret := os.Getenv("NOTION_SECRET")
	if len(notionSecret) == 0 {
		panic("NOTION_SECRET is not set")
	}
	databaseID := os.Getenv("NOTION_DATABASE_ID")
	if len(databaseID) == 0 {
		panic("NOTION_DATABASE_ID is not set")
	}

	storage := NewUserStateStorage()
	manager := NewNotionManager(notion.NewClient(notionSecret), databaseID)

	tgManager := NewTelegramManager(bot, offset, 30, 2, false, storage, manager)
	if err := tgManager.ListenAndServe(ctx); err != nil {
		panic(err)
	}

}

func GetLogger() *slog.Logger {
	once.Do(func() {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	})
	return log
}
