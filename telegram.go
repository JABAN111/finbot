package main

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"sync"
)

var (
	errNotCommand = errors.New("this msg is not a command")
)

const (
	commandStart = "start"
	commandHelp  = "help"
	commandReset = "reset"
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
						tm.log.Error("updates channel is closed")
						return
					}

					currentUserState, exist := users[update.Message.From.ID]
					if !exist {
						currentUserState = UserState{}
						users[update.Message.From.ID] = currentUserState
					}

					if update.Message != nil {
						if err := tm.processMessage(update); err != nil {
							if errors.Is(err, errNotCommand) {
								continue
							}
							tm.log.Error("fail to process command", "error", err, "user_id", update.Message.From.ID)
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
		Command:     commandStart,
		Description: "Начать работу с ботом",
	}, tgbotapi.BotCommand{
		Command:     commandHelp,
		Description: "Запросить описание взаимодействия с ботом",
	}, tgbotapi.BotCommand{
		Command:     commandReset,
		Description: "Сбрасывает состояние заполнение данных для операции",
	})

	_, err := tm.bot.Request(cfg)
	return err
}

func (tm *TelegramManager) processMessage(update tgbotapi.Update) error {
	msg := update.Message
	switch msg.Command() {
	case commandStart:
		tm.log.Info("Received /start command", "user_id", msg.From.ID)
		return nil
	case commandHelp:
		tm.log.Info("Received /help command", "user_id", msg.From.ID)
		helpText := "Здесь будет текст для команды /help"
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, helpText)
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

	case commandReset:
		tm.log.Info("Received /reset command", "user_id", msg.From.ID)
		users[update.Message.Chat.ID] = UserState{}
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Статус заполнения данных для операции сброшен")
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

	default:
		tm.log.Warn("Received unknown command", "command", msg.Command(), "user_id", msg.From.ID)
		return errNotCommand
	}
	return nil
}
