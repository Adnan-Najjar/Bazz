package bot

import (
	"math"
	"log"
	"math/rand"
	"net/http"
)

func Max(slice []float64) float64 {
	if len(slice) == 0 {
		return math.NaN()
	}

	maxValue := slice[0]
	for _, value := range slice {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func Min(slice []float64) float64 {
	if len(slice) == 0 {
		return math.NaN()
	}

	maxValue := slice[0]
	for _, value := range slice {
		if value < maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func randGet(url string) (*http.Response, error) {
	var userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.1 Safari/605.1.15",
		"Mozilla/5.0 (Linux; Android 10; Pixel 3 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Gecko/20100101 Firefox/89.0",
	}
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request with a random User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	// Perform the GET request
	resp, err := client.Do(req)
	return resp, err
}
