package main

import (
	"context"
	"fmt"
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
	ColorGray   = Color("Gray")
	ColorBrown  = Color("Brown")
	ColorOrange = Color("Orange")
	ColorYellow = Color("Yellow")
	ColorGreen  = Color("Green")
	ColorBlue   = Color("Blue")
	ColorPurple = Color("Purple")
	ColorPink   = Color("Pink")
	ColorRed    = Color("Red")
)

const (
	ColumnWho           = "Кто"
	ColumnOperationSum  = "Сумма операции"
	ColumnComment       = "Комментарий"
	ColumnCategory      = "Категория"
	ColumnStatus        = "Статус"
	ColumnOperationDate = "Дата внесения"
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

type InsertOperationData struct {
	OperationCreator  string
	OperationCategory string
	OperationSum      float64
	OperationStatus
}

func (n *NotionManager) PrintPage(ctx context.Context, pageID string) error {
	page, err := n.client.FindPageByID(ctx, pageID)
	if err != nil {
		n.log.Error("fail to find a page by id", "err", err, "page id", pageID)
		return err
	}

	fmt.Printf("found page: %v\n", page)
	return nil
}

func (n *NotionManager) QueryDatabase() {
	databaseID := "21d3d8e126bf8022a26dc42a29697985"
	ctx := context.Background()

	query := &notion.DatabaseQuery{}
	resp, err := n.client.QueryDatabase(ctx, databaseID, query)
	if err != nil {
		n.log.Error("ошибка при запросе базы данных", "err", err)
		panic(err)
	}
	for _, p := range resp.Results {
		fmt.Printf("got page: %v\n", p)

		switch props := p.Properties.(type) {
		case notion.PageProperties:
			for _, rt := range props.Title.Title {
				fmt.Println("Заголовок:", rt.Text.Content)
			}
		case notion.DatabasePageProperties:
			for name, prop := range props {
				fmt.Println("name", name)
				value := prop.Value()
				fmt.Println(value)
			}
		default:
			log.Error(fmt.Sprintf("неожиданный тип Properties: %T", props))
			panic(err)
		}

	}
}

func (n *NotionManager) InsertToDb(ctx context.Context) {
	databaseID := "21d3d8e126bf8022a26dc42a29697985"

	props := notion.DatabasePageProperties{
		ColumnWho: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name: "Миша",
			},
		},
		ColumnCategory: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name:  "Продуктывdsamklыф",
				Color: "pu",
			},
		},
		ColumnComment: notion.DatabasePageProperty{
			Title: []notion.RichText{
				{
					Text: &notion.Text{Content: "на июнь месяц"},
				},
			},
		},
		ColumnOperationSum: notion.DatabasePageProperty{
			Number: notion.Float64Ptr(-32918),
		},
		ColumnOperationDate: notion.DatabasePageProperty{
			Date: &notion.Date{
				Start: notion.NewDateTime(time.Now(), false),
			},
		},
		ColumnStatus: notion.DatabasePageProperty{
			Status: &notion.SelectOptions{
				Name: "Снятие",
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
		log.Error(fmt.Sprintf("Ошибка создания страницы: %v", err))
		return
	}
	log.Info(fmt.Sprintf("Страница создана: %s", page.URL))
}
