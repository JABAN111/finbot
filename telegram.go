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
	errMessageIsNil   = errors.New("message is nil")
)

const (
	commandStart = "start"
	commandHelp  = "help"
	commandReset = "reset"
)

type TelegramManager struct {
	numWorkers   int
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	log          *slog.Logger
	mu           sync.Mutex

	users   map[int64]UserState
	actions map[string]ButtonAction
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout, numWorkers int, debug bool) *TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	mapButtonActions := map[string]ButtonAction{buttonRefill: RefillButtonAction{}, buttonRemove: RemoveButtonAction{}}

	tm := &TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
		users:      make(map[int64]UserState),
		mu:         sync.Mutex{},
		actions:    mapButtonActions,
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

					tm.updateUserState(update)
					if update.CallbackQuery != nil {
						if err := tm.processCallbackQuery(update); err != nil {
							//tm.log.Error("fail to process callback query", "error", err)
						}
					}

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

	userState := tm.getCurrentUserState(msg.From.ID)

	if !msg.IsCommand() && !userState.isWaitUserInput {
		tm.log.Info("message is not command", "user_id", userState, "text", msg.Text)
		response := tgbotapi.NewMessage(update.Message.Chat.ID, "все хиханьки хаханьки тебе? Жмякни по /start")
		if _, err := tm.bot.Send(response); err != nil {
			return err
		}

		return nil
	}

	switch msg.Command() {
	case commandStart:
		tm.log.Info("Received /start command", "user_id", msg.From.ID)
		keyboard := createMainInlineCommands()
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
		tm.users[update.Message.Chat.ID] = UserState{}
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

func (tm *TelegramManager) processCallbackQuery(update tgbotapi.Update) error {
	msg := update.CallbackQuery
	//chatID := msg.Chat.ID
	userID := msg.From.ID

	tm.mu.Lock()
	currentUserState, ok := tm.users[userID]
	if !ok {
		tm.log.Error("user missed in users map", "userID", userID)
		currentUserState = UserState{}
		tm.users[userID] = currentUserState
	}
	tm.mu.Unlock()

	//switch update.CallbackQuery.Data {
	//case buttonRefill:
	//	keyboard := tm.createRefillKeys()
	//	response := tgbotapi.NewMessage(userID, "ты шо ебанутый?")
	//	response.ReplyMarkup = keyboard
	//	if _, err := tm.bot.Send(response); err != nil {
	//		return err
	//	}
	//case buttonRemove:
	//	response := tgbotapi.NewMessage(userID, "удоляемся")
	//	if _, err := tm.bot.Send(response); err != nil {
	//		return err
	//	}
	//case buttonBackToMain:
	//	keyboard := tm.createMainInlineCommands()
	//	responseMsg := tgbotapi.NewMessage(userID, "Выберите действие")
	//	responseMsg.ReplyMarkup = keyboard
	//	_, err := tm.bot.Send(responseMsg)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (tm *TelegramManager) getCurrentUserState(userID int64) UserState {
	usState, ok := tm.users[userID]

	if !ok {
		tm.mu.Lock()
		tm.users[userID] = UserState{}
		tm.mu.Unlock()
	}

	return usState
}
