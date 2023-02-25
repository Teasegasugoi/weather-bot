package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

var (
	SLACK_TOKEN, CHANNEL_ID string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}
	SLACK_TOKEN = os.Getenv("BOT_USER_OAUTH_TOKEN")
	CHANNEL_ID = os.Getenv("CHANNEL_ID")
}

func main() {
	c := slack.New(SLACK_TOKEN)
	// MsgOptionText() の第二引数に true を設定すると特殊文字をエスケープする
	_, _, err := c.PostMessage(CHANNEL_ID, slack.MsgOptionText("Hello World", true))
	if err != nil {
		panic(err)
	}
}
