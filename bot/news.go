package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron"
)

type Events struct {
	Sentiment string `json:"sentiment"`
	Ticker    string `json:"flagCur"`
	Event     string `json:"event"`
	Previous  string `json:"prev"`
	Forecast  string `json:"fore"`
	Actual    string `json:"act"`
}

func AnalyisNews() (string, error) {
	urls := []string{
		// Business
		"https://news.google.com/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx6TVdZU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US%3Aen",
		// World
		"https://news.google.com/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx1YlY4U0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US%3Aen",
		// US
		"https://news.google.com/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx1YlY4U0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US%3Aen",
	}

	// Shuffle the urls slice
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	headlines := ""
	for x := range urls {
		url := urls[x]
		headline := googleNews(url)

		log.Printf("\nLoading url: %s", url)

		// Sleeps a random interval
		time.Sleep(time.Duration(rand.Intn(7)) * time.Second)

		headlines += headline
	}

	log.Printf("Loading Done!")
	prompt := "**Headlines:** " + headlines

	settings := Settings{
		Tempreture: 0.5,
		TopP:       0.5,
		TopK:       10,
	}

	system := readFile("analyis.md")

	log.Printf("Waiting for AI response...")
	res, err := AiResponse(prompt, system, settings)
	log.Println(res)
	return res, err
}

// read system files
func readFile(fileName string) string {
	filePath := filepath.Join(BaseDir, "system", fileName)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Cannot open file: %s", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Cannot read file: %s", err)
	}
	return string(fileContent)
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

func googleNews(url string) string {
	// Make an HTTP GET request to the website
	resp, err := randGet(url)
	if err != nil {
		log.Println("Error making request:", err)
	}
	defer resp.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Parsing data...")
	all := ""
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		title := s.Text()
		if len(title) > 40 {
			all += title + "\n"
		}
	})
	return all
}

func investNews(wg *sync.WaitGroup) {
	defer wg.Done()
	url := "https://sslecal2.investing.com/?columns=exc_flags,exc_currency,exc_importance,exc_actual,exc_forecast,exc_previous&category=_employment,_economicActivity,_inflation,_credit,_centralBanks,_confidenceIndex,_balance,_Bonds&importance=3&countries=6,37,72,35,43,56,4,5&calType=week&timeZone=8&lang=3"
	log.Println(url)
	res, err := randGet(url)
	if err != nil {
		log.Println("Investing.com request errored: ", err)
	}
	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Parsing data...")

	extractedClasses := make(map[string]Events)

	count := 0
	doc.Find("tr").Each(func(_ int, s *goquery.Selection) {
		var event Events
		if timestamp, exist := s.Attr("event_timestamp"); exist {
			for _, className := range []string{"flagCur", "event", "prev", "fore", "act"} {
				data := s.Find("td." + className).Text()
				data = strings.ReplaceAll(data, "\u00A0", "0")
				data = strings.TrimSpace(data)
				switch className {
				case "flagCur":
					event.Ticker = strings.TrimSpace(strings.ReplaceAll(data, "0", ""))
				case "event":
					event.Event = strings.TrimSpace(strings.ReplaceAll(data, "0", ""))
				case "prev":
					event.Previous = data
				case "fore":
					event.Forecast = data
				case "act":
					event.Actual = data
				}
			}
			extractedClasses[fmt.Sprintf("%03d", count)+timestamp] = event
			count++
		}
	})

	// Convert to JSON
	jsonData, err := json.MarshalIndent(extractedClasses, "", "  ")
	if err != nil {
		log.Println("Error marshalling to JSON:", err)
		return
	}

	// Save JSON data to a file
	err = os.WriteFile("economic-calendar.json", jsonData, 0644)
	if err != nil {
		log.Println("Error writing to file:", err)
		return
	}

	log.Println("JSON data saved to economic-calendar.json")
}

func ScheduleAlert() {
	file, err := os.Open("economic-calendar.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var events map[string]Events
	if err := json.NewDecoder(file).Decode(&events); err != nil {
		log.Fatalf("Error in reading JSON! %s", err)
	}

	// exec before market open
	todayTimes := []string{""}
	currentTime := time.Now().UTC().Format("2006-01-02")
	if len(todayTimes) < 2 || todayTimes[0] != currentTime {
		todayTimes[0] = currentTime
		for event := range events {
			if eventTime, exist := strings.CutPrefix(event[2:], currentTime); exist {
				// fmt.Println(events[event]) // send embed message of them all
				eventTime = strings.TrimSpace(eventTime)
				todayTimes = append(todayTimes, eventTime)
			}
		}
	}

	scheduler := gocron.NewScheduler(time.UTC)
	for _, timeStr := range todayTimes[1:] {
		scheduler.At(timeStr).Do(func() {
			fmt.Println("YESS!")
		})
	}

	// Start the scheduler
	scheduler.StartAsync()
}
