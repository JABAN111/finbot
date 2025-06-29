package main

import (
	"context"
	"github.com/dstotijn/go-notion"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"testing"
)

func TestNotionManager_PrintPage(t *testing.T) {
	type fields struct {
		log    *slog.Logger
		client *notion.Client
	}
	type args struct {
		ctx    context.Context
		pageID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "attempt to get a page",
			fields: fields{
				log:    GetLogger(),
				client: getNotionClient(),
			},
			args: args{
				ctx:    context.Background(),
				pageID: "21d3d8e126bf8096ae11f9a4d100b714",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NotionManager{
				log:    tt.fields.log,
				client: tt.fields.client,
			}
			if err := n.PrintPage(tt.args.ctx, tt.args.pageID); (err != nil) != tt.wantErr {
				t.Errorf("PrintPage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getNotionClient() *notion.Client {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	token := os.Getenv("NOTION_SECRET")

	return notion.NewClient(token)
}

func TestNotionManager_QueryDatabase(t *testing.T) {
	c := getNotionClient()
	nm := NewNotionManager(c)

	nm.QueryDatabase()
}

func TestNotionManager_InsertToDb(t *testing.T) {

	c := getNotionClient()
	nm := NewNotionManager(c)

	nm.InsertToDb(context.Background())

}
