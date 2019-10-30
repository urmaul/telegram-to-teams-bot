package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/mkideal/cli"
	"github.com/sirupsen/logrus"
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

func getMessageText(message *tgbotapi.Message) string {
	text := message.Text
	if message.Photo != nil {
		text = "(photo)"
	}
	if message.Video != nil {
		text = "(video)"
	}
	if message.Audio != nil {
		text = "(audio)"
	}
	if message.Sticker != nil {
		text = fmt.Sprintf("(%s sticker)", message.Sticker.Emoji)
	}

	return text
}

type argT struct {
	cli.Helper
	TelegramToken           string         `cli:"*telegram-token" usage:"telegram bot token" dft:"$TELEGRAM_TOKEN"`
	TelegramChatID          int64          `cli:"*telegram-chat-id" usage:"id of telegram chat to forward" dft:"$TELEGRAM_CHAT_ID"`
	MSTeamsWebhookURL       string         `cli:"*msteams-webhook-url" usage:"webhook url to post to msteams channel" dft:"$MSTEAMS_WEBHOOK_URL"`
	MSTeamsPersonalWebhooks map[int]string `cli:"W" usage:"personal webhooks per telegram user id" dft:"$MSTEAMS_PERSONAL_WEBHOOKS"`
	Log                     string         `cli:"log" usage:"log level"`
}

func main() {
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		logger := logrus.New()

		logLevel, err := logrus.ParseLevel(argv.Log)
		if err != nil {
			logger.Fatalf(err.Error())
		}
		logger.SetLevel(logLevel)

		defaultWebhookURL := argv.MSTeamsWebhookURL
		personalWebhookURLs := argv.MSTeamsPersonalWebhooks

		chatID := argv.TelegramChatID
		logger.Debugf("Telegram chat ID: %d", chatID)
		logger.Debugf("Personal webhooks: %+v", personalWebhookURLs)

		bot, err := tgbotapi.NewBotAPI(argv.TelegramToken)
		if err != nil {
			logger.Fatalf("Error when trying to connect to Telegram: %s", err)
		}

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

			text := getMessageText(update.Message)
			if text == "" {
				logger.Errorf("Could not get body for message %+v", update.Message)
				continue
			}

			webhookURL, found := personalWebhookURLs[update.Message.From.ID]
			msg := text
			if !found {
				// Use default webhook, provide name inside message
				msg = fmt.Sprintf("@%s: %s", update.Message.From.UserName, text)
				webhookURL = defaultWebhookURL
			}

			// Submit

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
