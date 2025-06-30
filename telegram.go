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
	errUnknownCommand = errors.New("get unknown command")
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
	mu           sync.RWMutex
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

					tm.mu.Lock()
					currentUserState, exist := users[update.Message.From.ID]
					if !exist {
						currentUserState = UserState{}
						users[update.Message.From.ID] = currentUserState
					}
					tm.mu.Unlock()

					if update.Message != nil {
						if err := tm.processMessage(update); err != nil {
							if errors.Is(err, errUnknownCommand) {
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

	if update.CallbackQuery != nil {
		return tm.handleCallback(update)
	}

	switch msg.Command() {
	case commandStart:
		tm.log.Info("Received /start command", "user_id", msg.From.ID)
		keyboard := tm.createMainInlineCommands()
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Выберите действие")
		responseMsg.ReplyMarkup = keyboard
		_, err := tm.bot.Send(responseMsg)
		if err != nil {
			return err
		}
		return nil
	case commandHelp:
		tm.log.Info("Received /help command", "user_id", msg.From.ID)
		helpText := "Здесь будет текст для команды /help"
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, helpText)
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

	case commandReset:
		tm.mu.Lock()
		tm.log.Info("Received /reset command", "user_id", msg.From.ID)
		users[update.Message.Chat.ID] = UserState{}
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Статус заполнения данных для операции сброшен")
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}
		tm.mu.Unlock()

	default:
		tm.log.Warn("Received unknown command", "command", msg.Command(), "user_id", msg.From.ID)
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Неизвестная команда: %s", msg.Command()))
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}
		return errUnknownCommand
	}

	return nil
}

func (tm *TelegramManager) createMainInlineCommands() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пополнение", "refill"),
			tgbotapi.NewInlineKeyboardButtonData("Снятие", "remove"),
		),
	)
}

func (tm *TelegramManager) createChooseCategoryInlineCommands() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Продукты", "category_products"),
			tgbotapi.NewInlineKeyboardButtonData("Транспорт", "category_transport"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Развлечения", "category_entertainment"),
			tgbotapi.NewInlineKeyboardButtonData("Коммуналка", "category_utility"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_main"),
		),
	)
}

func (tm *TelegramManager) handleCallback(update tgbotapi.Update) error {
	msg := update.Message
	//chatID := msg.Chat.ID
	userID := msg.From.ID

	tm.mu.Lock()
	currentUserState, ok := users[userID]
	if !ok {
		tm.log.Error("user missed in users map", "userID", userID)
		currentUserState = UserState{}
		users[userID] = currentUserState
	}
	tm.mu.Unlock()

	return nil
}
