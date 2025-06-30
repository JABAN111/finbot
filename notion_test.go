package main

import (
	"context"
	"github.com/dstotijn/go-notion"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func getNotionClient() *notion.Client {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	token := os.Getenv("NOTION_SECRET")
	if len(token) == 0 {
		panic("NOTION_SECRET is not set")
	}
	return notion.NewClient(token)
}

func TestNotionManager_InsertOperation(t *testing.T) {
	err := godotenv.Load()
	require.NoError(t, err)
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	ctx := context.Background()

	tests := []struct {
		name    string
		dto     InsertOperationDto
		wantErr bool
	}{
		{
			name: "fully valid insert",
			dto: InsertOperationDto{
				Creator:  "jaba368",
				Category: "продукты",
				Sum:      38218321,
				Status:   OperationStatusRefill,
				Comment:  "Обычное добавление после теста",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notionClient := getNotionClient()
			nm := NewNotionManager(notionClient)
			if err := nm.InsertOperation(ctx, databaseID, tt.dto); (err != nil) != tt.wantErr {
				t.Errorf("InsertOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
