package main

import (
	"context"
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
	tgManager := NewTelegramManager(bot, offset, 30, 2, false)
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
