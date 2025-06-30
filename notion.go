package main

import (
	"context"
	"github.com/dstotijn/go-notion"
	"log/slog"
	"time"
)

type Color string
type OperationStatus string

const (
	OperationStatusRefill = OperationStatus("Пополнение")
	OperationRemove       = OperationStatus("Снятие")
)

const (
	columnWho           = "Кто"
	columnOperationSum  = "Сумма операции"
	columnComment       = "Комментарий"
	columnCategory      = "Категория"
	columnStatus        = "Статус"
	columnOperationDate = "Дата внесения"
)

type NotionManager struct {
	log    *slog.Logger
	client *notion.Client
}

func NewNotionManager(client *notion.Client) NotionManager {
	return NotionManager{
		log:    GetLogger(),
		client: client,
	}
}

type InsertOperationDto struct {
	Creator  string
	Category string
	Sum      float64
	Status   OperationStatus
	Comment  string
}

func (n NotionManager) InsertOperation(ctx context.Context, databaseID string, operationDTO InsertOperationDto) error {
	props := notion.DatabasePageProperties{
		columnWho: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name: operationDTO.Creator,
			},
		},
		columnCategory: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name: operationDTO.Category,
			},
		},
		columnComment: notion.DatabasePageProperty{
			Title: []notion.RichText{
				{
					Text: &notion.Text{Content: operationDTO.Comment},
				},
			},
		},
		columnOperationSum: notion.DatabasePageProperty{
			Number: &operationDTO.Sum,
		},
		columnOperationDate: notion.DatabasePageProperty{
			Date: &notion.Date{
				Start: notion.NewDateTime(time.Now(), false),
			},
		},
		columnStatus: notion.DatabasePageProperty{
			Status: &notion.SelectOptions{
				Name: string(operationDTO.Status),
			},
		},
	}

	params := notion.CreatePageParams{
		ParentType:             notion.ParentTypeDatabase,
		ParentID:               databaseID,
		DatabasePageProperties: &props,
	}

	page, err := n.client.CreatePage(ctx, params)
	if err != nil {
		log.Error("Ошибка создания страницы", "err", err)
		return err
	}
	log.Info("Страница создана", "url", page.URL)
	return nil
}
