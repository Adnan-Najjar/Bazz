package bot

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Getting chart image using chart-img.com api
func GetChart(wg *sync.WaitGroup, state string, symbol string, sl float64, entryLow float64, entryHigh float64, lot float64, maxTp float64) {

	defer wg.Done()
	url := "https://api.chart-img.com/v2/tradingview/advanced-chart/"
	apiKey, _ := os.LookupEnv("CHART_API")

	// Getting time in DateTime (ISO8601) for the API
	startTime := time.Now().UTC().Format(time.RFC3339)

	data := map[string]interface{}{
		"height":   600,
		"width":    800,
		"theme":    "dark",
		"interval": "15m",
		"symbol":   symbol,
		"drawings": []interface{}{
			// Long Position for Stop Loss and Take Profit Area
			map[string]interface{}{
				"name": "Long Position",
				"input": map[string]interface{}{
					"startDatetime": startTime,
					"entryPrice":    entryLow,
					"targetPrice":   maxTp,
					"stopPrice":     sl,
				},
				"override": map[string]interface{}{
					"showStats":   false,
					"lotSize":     lot,
					"showCompact": true,
				},
			},
			// Entry Area using Long Position
			map[string]interface{}{
				"name": "Long Position",
				"input": map[string]interface{}{
					"startDatetime": startTime,
					"entryPrice":    entryLow,
					"targetPrice":   entryHigh,
					"stopPrice":     entryHigh + 10,
				},
				"override": map[string]interface{}{
					"lineColor":        "rgb(0,255,255)",
					"borderColor":      "rgb(0,255,255)",
					"profitBackground": "rgba(0,255,255,0.2)",
					"stopBackground":   "rgba(0,0,0,0)",
				},
				"zOrder": "top",
			},
			// Stop Loss line
			map[string]interface{}{
				"name": "Horizontal Line",
				"input": map[string]interface{}{
					"text":  "Stop Loss",
					"price": sl,
				},
				"override": map[string]interface{}{
					"lineColor":      "rgb(255,0,0)",
					"textColor":      "rgb(255,0,0)",
					"horzLabelAlign": "right",
					"showPrice":      false,
				},
			},
		},
	}

	// Marshal the map into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Sending the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Create the output file
	filepath := filepath.Join(BaseDir, "chart", "chart-img.png")
	out, err := os.Create(filepath)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	// Write the image to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("Error saving response to file:", err)
		return
	}

}
