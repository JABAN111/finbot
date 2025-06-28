package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

const offset = 0 //??

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := godotenv.Load()
	if err != nil {
		log.Warn("fail to read from .env file", "err", err)
	}

	token := os.Getenv("TELEGRAM_SECRET")
	if len(token) == 0 {
		log.Error("telegram secret is not specified")
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	bot.Debug = true
	if err != nil {
		log.Error("fail to create bot", "error", err)
		os.Exit(1)
	}

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = 30

	updatesChan := bot.GetUpdatesChan(tgConfig)
	for data := range updatesChan {
		if data.Message == nil {
			continue
		}

		log.Info("got message", "message", data.Message)
		fmt.Printf("data struct: %v\n", data)
	}
}
