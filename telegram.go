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

	tm := TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
	}

	err := tm.setCommands()
	if err != nil {
		log.Error("fail to set commands for telegram bot", "error", err)
		panic(fmt.Errorf("fail to set commands for telegram bot: %e", err))
	}
	return tm
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
					fmt.Println("Received update:", update)

					if msg := update.Message; msg != nil {
						switch msg.Command() {
						case "start":
							tm.log.Info("Received /start command", "user_id", msg.From.ID)
							// Handle /start command

						case "help":
							tm.log.Info("Received /help command", "user_id", msg.From.ID)

						case "reset":
							tm.log.Info("Received /reset command", "user_id", msg.From.ID)
						}
					}

				}
			}
		}()
	}

	wg.Wait()
	return nil
}

func (tm *TelegramManager) setCommands() error {
	cfg := tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{
		Command:     "start",
		Description: "Начать работу с ботом",
	}, tgbotapi.BotCommand{
		Command:     "help",
		Description: "Запросить описание взаимодействия с ботом",
	}, tgbotapi.BotCommand{
		Command:     "reset",
		Description: "Сбрасывает состояние заполнение данных для операции",
	})

	_, err := tm.bot.Request(cfg)
	return err
}
