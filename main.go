package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type msteamsMessage struct {
	Text string `json:"text"`
}

func pushToMsteams(msg string, webhookURL string) {
	message := msteamsMessage{Text: msg}
	messageString, err := json.Marshal(message)
	if err != nil {
		log.Print(err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageString))
	if err != nil {
		log.Print(err)
	}
	if resp.StatusCode != 200 {
		log.Printf("Got response code %d when sending message to msteams", resp.StatusCode)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	webhookURL := os.Getenv("MSTEAMS_WEBHOOK_URL")

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		fmt.Printf("update: %+v\n", update)
		if update.Message == nil { // Ignore non-Message updates
			continue
		}

		if update.Message.Chat == nil || update.Message.Chat.ID != -999999 { // Forward only messages from selected chat
			continue
		}

		msg := fmt.Sprintf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
		go pushToMsteams(msg, webhookURL)

		log.Println(msg)
	}
}
