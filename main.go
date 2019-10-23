package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/alexcesaro/log/stdlog"
	"github.com/mkideal/cli"
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

type argT struct {
	cli.Helper
	TelegramToken     string `cli:"*telegram-token" usage:"telegram bot token" dft:"$TELEGRAM_TOKEN"`
	TelegramChatID    int64  `cli:"*telegram-chat-id" usage:"id of telegram chat to forward" dft:"$TELEGRAM_CHAT_ID"`
	MSTeamsWebhookURL string `cli:"*msteams-webhook-url" usage:"webhook url to post to msteams channel" dft:"$MSTEAMS_WEBHOOK_URL"`
	Log               string `cli:"log" usage:"log level"`
}

func main() {
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		logger := stdlog.GetFromFlags()

		bot, err := tgbotapi.NewBotAPI(argv.TelegramToken)
		if err != nil {
			log.Fatalf("Error when trying to connect to Telegram: %s", err)
		}

		webhookURL := argv.MSTeamsWebhookURL

		chatID := argv.TelegramChatID
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

			logger.Debugf("Message is: %+v", update.Message)

			if update.Message.Chat == nil || update.Message.Chat.ID != chatID { // Forward only messages from selected chat
				logger.Debugf("Ignoring message from chat %+v", update.Message.Chat)
				continue
			}

			// Build message text

			text := update.Message.Text
			if update.Message.Photo != nil {
				text = "(photo)"
			}
			if update.Message.Video != nil {
				text = "(video)"
			}
			if update.Message.Audio != nil {
				text = "(audio)"
			}
			if update.Message.Sticker != nil {
				text = fmt.Sprintf("(%s sticker)", update.Message.Sticker.Emoji)
			}

			if text == "" {
				logger.Errorf("Could not get body for message %+v", update.Message)
				continue
			}

			msg := fmt.Sprintf("@%s: %s", update.Message.From.UserName, text)

			go func() {
				logger.Debugf("Sending message: %s", msg)
				err := pushToMsteams(msg, webhookURL)
				if err != nil {
					logger.Error(err)
				}
			}()
		}

		return nil
	}))
}
