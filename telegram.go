package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"sync"
)

var users = make(map[int64]UserState)

type TelegramManager struct {
	numWorkers   int
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	log          *slog.Logger
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout, numWorkers int, debug bool) TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	return TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
	}
}

func (tm *TelegramManager) ListenAndServe(ctx context.Context) error {
	upChan := tm.bot.GetUpdatesChan(tm.updateConfig)

	var wg sync.WaitGroup
	wg.Add(tm.numWorkers)

	for i := 0; i < tm.numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					tm.log.Warn("context for telegram is done")
					return
				case update, ok := <-upChan:
					if !ok {
						return
					}

					if update.CallbackQuery != nil {
						chatID := update.CallbackQuery.ID
						fmt.Println("Received update from chat ID:", chatID)
					}

				}
			}
		}()
	}

	wg.Wait()
	return nil
}
