package bot

import (
	"bytes"
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	botToken, _       = os.LookupEnv("TELE_API")
	chatID      int64 = -4506920657
)

// createBot initializes a new Telegram bot instance.
func createBot() *bot.Bot {
	b, err := bot.New(botToken)
	if err != nil {
		log.Printf("Failed to create bot: %v", err)
		return nil
	}
	return b
}

// TelegramSendMessage sends a message to a Telegram chat.
func TelegramSendMessage(text string, isMarkdown bool) {
	b := createBot()
	if b == nil {
		return
	}

	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}

	if isMarkdown {
		params.ParseMode = "MarkdownV2"
	}

	if _, err := b.SendMessage(context.Background(), params); err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	log.Println("Telegram message sent successfully!")
}

// TelegramSendPhoto sends a photo to a Telegram chat.
func TelegramSendPhoto(filePath string) {
	b := createBot()
	if b == nil {
		return
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", filePath, err)
		return
	}

	params := &bot.SendPhotoParams{
		ChatID: chatID,
		Photo:  &models.InputFileUpload{Filename: filePath, Data: bytes.NewReader(fileContent)},
	}

	if _, err := b.SendPhoto(context.Background(), params); err != nil {
		log.Printf("Failed to send photo: %v", err)
		return
	}

	log.Println("Telegram photo sent successfully!")
}

// Simple Make Recommendation for Telegram or any other platform
func AddRec(warning bool, state string, symbol string, sl float64, entryLow float64, entryHigh float64, tps string, description string, lot float64) string {
	recString := `
%s
**%s %s**
**Entry**: ➡️ __**%s**__ - __**%s**__

SL ❌ ~~%s~~ ❌
%s
%s

دخول ب **%s** لكل 1,000$ من رأس المال
`
	var warn string
	if warning {
		warn = "__**للمحافظ الكبيرة فقط**__⚠️"
	} else {
		warn = ""
	}

	if strings.ToUpper(state) == "SELL" {
		state = "SELL!🔴🔴🔴"
	} else {
		state = "BUY!🔵🔵🔵"
	}

	return fmt.Sprintf(recString, warn, symbol, state, strconv.FormatFloat(entryLow, 'g', -1, 64), strconv.FormatFloat(entryHigh, 'g', -1, 64), strconv.FormatFloat(sl, 'g', -1, 64), tps, description, strconv.FormatFloat(lot, 'g', -1, 64))
}
