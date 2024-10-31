package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var chatID string = "-4506920657"
var botToken, _ = os.LookupEnv("TELE_API")

func TelegramSendMessage(wg *sync.WaitGroup, text string) {
	defer wg.Done()
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Create the request payload
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "MarkdownV2",
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("\nFailed to send message: %s", resp.Status)
	}
}

// func TelegramSendPhoto(wg *sync.WaitGroup){
// 	defer wg.Done()
// 	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", botToken)
//
// 	// Create the request payload
// 	payload := map[string]interface{}{
// 		"chat_id":    chatID,
// 	}
// }


// Simple Make Recommendation for Telegram or any other platform
func AddRec(warning bool, state string, symbol string, sl float64, entryLow float64, entryHigh float64, tps string, description string, lot float32) string {
	recString := `
%s
**%s %s**
**Entry**:➡️ __**%.3f**__-__**%.3f**__

SL❌~~%.3f~~❌
%s
%s

دخول ب  **%.3f** لكل 1,000$ من رأس المال
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

	return fmt.Sprintf(recString, warn, symbol, state, entryLow, entryHigh, sl, tps, description, lot)
}
