package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/alexcesaro/log/stdlog"
)

type msteamsMessage struct {
	Text string `json:"text"`
}

func pushToMsteams(msg string, webhookURL string) error {
	message := msteamsMessage{Text: msg}
	messageString, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageString))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Got response code %d when sending message to msteams", resp.StatusCode)
	}

	return nil
}

func main() {
	logger := stdlog.GetFromFlags()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	webhookURL := os.Getenv("MSTEAMS_WEBHOOK_URL")

	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		log.Panic(err)
	}
	logger.Debugf("Telegram chat ID: %d", chatID)

	logger.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		logger.Debugf("Got update: %+v", update)

		if update.Message == nil { // Ignore non-Message updates
			continue
		}

		if update.Message.Chat == nil || update.Message.Chat.ID != chatID { // Forward only messages from selected chat
			continue
		}

		msg := fmt.Sprintf("@%s: %s", update.Message.From.UserName, update.Message.Text)

		go func() {
			logger.Debugf("Sending message: %s", msg)
			err := pushToMsteams(msg, webhookURL)
			if err != nil {
				logger.Error(err)
			}
		}()
	}
}
