package main

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"strconv"
	"strings"
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

	storage Storage
	actions map[string]ButtonAction
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout, numWorkers int, debug bool, storage Storage) *TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	mapButtonActions := map[string]ButtonAction{
		buttonRefill:     RefillButtonAction{storage: storage},
		buttonRemove:     RemoveButtonAction{storage: storage},
		buttonBackToMain: ReturnToMainButtonAction{storage: storage},
	}

	tm := &TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
		storage:    storage,
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

					if update.CallbackQuery != nil {
						if err := tm.processCallbackQuery(update); err != nil {
							//tm.log.Error("fail to process callback query", "error", err)
						}
					}

					if update.Message == nil {
						continue
					}

					if err := tm.processMessage(update); err != nil {
						if errors.Is(err, errUnknownCommand) {
							continue
						}
						tm.log.Error("fail to process command", "error", err, "user_id", update.Message.From.ID)
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
	if msg.IsCommand() {
		return tm.processCommand(update)
	}

	userState, err := tm.storage.Get(msg.From.ID)
	if err != nil {
		if !errors.Is(err, errUserNotFound) {
			tm.log.Error("fail to get user state", "error", err)
			return err
		}
		if err := tm.storage.Save(update.Message.From.ID, UserState{}); err != nil {
			return err
		}
		tm.log.Info("missed user created state", "user_id", update.Message.From.ID)
		userState = UserState{}
	}

	if !userState.isWaitUserInput {
		tm.log.Info("message is not command", "user_id", userState, "text", msg.Text)
		response := tgbotapi.NewMessage(update.Message.Chat.ID, "все хиханьки хаханьки тебе? Жмякни по /start")
		if _, err := tm.bot.Send(response); err != nil {
			return err
		}
		return nil
	}

	if userState.isWaitUserInput {
		userInput := msg.Text
		tm.log.Info("got expected input from user", "user_id", update.Message.From.ID, "userInput", userInput)

		switch userState.UserStateCurrentOperation {
		case settingSum:
			sum, err := strconv.ParseFloat(strings.TrimSpace(userInput), 64)
			if err != nil {
				_ = tm.sendSpecificErrMsg("не удалось прочесть сумму, пожалуйста введите в формате 0.321", update)
				return fmt.Errorf("fail to parse sum value: %w", err)
			}

			if sum < 0 {
				tm.log.Warn("user attempt to enter negative sum")
				_ = tm.sendSpecificErrMsg("сумма должна быть положительным числом", update)
				return errors.New("user attempt to enter negative sum")
			}

			userState.isWaitUserInput = false
			userState.OperationSum = sum
			if err := tm.storage.Save(update.Message.From.ID, userState); err != nil {
				tm.log.Error("fail to save user state", "error", err)
				return err
			}
			tm.log.Info("user entered sum", "user_id", update.Message.From.ID, "updated user state", userState)
			return nil
		}
	}

	return nil
}

func (tm *TelegramManager) processCommand(update tgbotapi.Update) error {
	msg := update.Message

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
		tm.log.Info("Received /reset command", "user_id", msg.From.ID)
		if err := tm.storage.Reset(update.Message.Chat.ID); err != nil {
			return err
		}
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Статус заполнения данных для операции сброшен")
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

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
	userID := msg.From.ID

	callbackData := update.CallbackQuery.Data
	button, ok := tm.actions[callbackData]

	if !ok {
		return nil
	}

	resp, err := button.Action(userID, update)
	if err != nil {
		return err
	}
	if _, err = tm.bot.Send(resp); err != nil {
		return err
	}
	tm.log.Info("sent message to user", "user_id", userID, "command", callbackData)

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

func (tm *TelegramManager) sendSpecificErrMsg(msgText string, update tgbotapi.Update) error {
	msg := update.Message
	userID := msg.From.ID

	errResp := tgbotapi.NewMessage(userID, msgText)
	if err := tm.storage.Reset(userID); err != nil {
		tm.log.Error("fail to reset user before sending error msg", "user_id", userID, "error", err)
		return err
	}

	if _, err := tm.bot.Send(errResp); err != nil {
		tm.log.Error("fail to send error msg", "user_id", userID, "error", err)
		return err
	}
	return nil
}

func (tm *TelegramManager) sendUnexpectedErrMsg(update tgbotapi.Update) error {
	return tm.sendSpecificErrMsg("произошла ошибка при обработке ваших ответов и/или внутренняя\nначните заново", update)
}
