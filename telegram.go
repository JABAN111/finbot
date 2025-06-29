package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

type TelegramManager struct {
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	log          *slog.Logger
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout int, debug bool) TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	return TelegramManager{
		bot: bot,
		log: GetLogger(),
	}
}

func (tm *TelegramManager) ListenAndServe() error {
	upChan := tm.bot.GetUpdatesChan(tm.updateConfig)
	for data := range upChan {
		if data.Message == nil {
			continue
		}

		tm.log.Info("got updates", "data", data)
	}

	return nil
}
