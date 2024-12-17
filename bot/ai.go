package bot

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
	"os"
)

type Settings struct {
	Tempreture float32
	TopP       float32
	TopK       int32
}

// Parse the response from Gemini
func getResponse(resp *genai.GenerateContentResponse) string {
	// Create a slice to hold the news
	var news string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for x := range cand.Content.Parts {
				news += fmt.Sprintln(cand.Content.Parts[x])
			}
		}
	}
	return news
}

// Takes a propmt and a system instruction to
// interact with the Gemini API and gets a response
func AiResponse(prompt string, system string, settings Settings) (string, error) {

	ctx := context.Background()
	// Access your API key as an environment variable
	api_key, _ := os.LookupEnv("AI_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(api_key))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(settings.Tempreture)
	model.SetTopP(settings.TopP)
	model.SetTopK(settings.TopK)
	model.SystemInstruction = genai.NewUserContent(genai.Text(system))
	model.ResponseMIMEType = "text/plain"

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	return getResponse(resp), err
}
