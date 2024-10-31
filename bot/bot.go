package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"unicode"

	"github.com/bwmarrin/discordgo"
)

var DgSession *discordgo.Session
var BaseDir string
var (
	IntegerOptionMinValue = 0.0

	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "news",
			Description: "Get latest news from google news and analyiz them",
		},
		{
			Name:        "mkrec",
			Description: "Makeing a Recommendation",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "Add description to the recommendation",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "warning",
					Description: "Is this for big money only?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "symbol",
					Description: "Choose from options or write it in this format EXCHANGE:SYMBOL (e.g. FX:GBPJPY)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "GOLD",
							Value: "OANDA:XAUUSD",
						},
						{
							Name:  "US100",
							Value: "CAPITALCOM:US100",
						},
						{
							Name:  "USOIL",
							Value: "TVC:USOIL",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "state",
					Description: "Choose buy or sell",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Buy",
							Value: "BUY",
						},
						{
							Name:  "Sell",
							Value: "SELL",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "stop-loss",
					Description: "Stop Loss",
					MinValue:    &IntegerOptionMinValue,
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "entry-low",
					Description: "First entry value",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "entry-high",
					Description: "Second entry value",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "take-profits",
					Description: "tp1 tp2 tp3 ...",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "lot-size",
					Description: "Lot Size",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "author",
					Description: "Add an author",
					Required:    false,
				},
			},
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Make Recommendation slash command
		"mkrec": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			mkrec := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(mkrec))
			for _, opt := range mkrec {
				optionMap[opt.Name] = opt
			}

			embed := &discordgo.MessageEmbed{
				Type: "rich",
			}

			// Adding the author
			if opt, ok := optionMap["author"]; ok {
				embed.Author = &discordgo.MessageEmbedAuthor{
					Name:    opt.UserValue(s).GlobalName,
					IconURL: opt.UserValue(s).AvatarURL("256"),
				}
			} else {
				embed.Author = &discordgo.MessageEmbedAuthor{
					Name:    i.Member.User.GlobalName,
					IconURL: i.Member.User.AvatarURL("256"),
				}
			}

			// Adding a warning message (if needed)
			var isWarn bool
			if opt, ok := optionMap["warning"]; ok {
				if isWarn = opt.BoolValue(); isWarn {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "__**⚠️ Warning**__",
						Value:  "__**للمحافظ الكبيرة فقط**__",
						Inline: false,
					})
				}
			}

			// Description at the start of the embed
			var description string
			if opt, ok := optionMap["description"]; ok {
				description = opt.StringValue()
				embed.Description = fmt.Sprintf("%s", description)
			}

			var tpFloats []float64
			var symbol string
			var state string
			var sl float64
			var entryLow float64
			var entryHigh float64
			var maxTp float64

			// Adding Entry Low & High
			if opt1, ok := optionMap["entry-low"]; ok {
				entryLow = opt1.FloatValue()
				if opt2, ok := optionMap["entry-high"]; ok {
					entryHigh = opt2.FloatValue()
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "Entry",
						Value:  fmt.Sprintf("➡️ __**%s**__ - __**%s**__", strconv.FormatFloat(entryLow, 'g', -1, 64), strconv.FormatFloat(entryHigh, 'g', -1, 64)),
						Inline: true,
					})
				}
			}

			// Adding Stop Loss
			if opt, ok := optionMap["stop-loss"]; ok {
				sl = opt.FloatValue()
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "~~Stop Loss (SL)~~",
					Value:  fmt.Sprintf("❌ __**%s**__ ❌", strconv.FormatFloat(sl, 'g', -1, 64)),
					Inline: true,
				})
			}

			// Adding Take profits
			// Making Tps slice
			var tpString string
			if opt, ok := optionMap["take-profits"]; ok {
				list := strings.Split(opt.StringValue(), " ")
				for i := range list {
					tp, _ := strconv.ParseFloat(list[i], 64)
					tpFloats = append(tpFloats, tp)
					tpString += fmt.Sprintf("*Tp%d ✅*	**%s**\n", i+1, strconv.FormatFloat(tp, 'g', -1, 64))
				}
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "__Take Profits (Tp)__",
					Value:  tpString,
					Inline: false,
				})

			}

			// Adding the ticker name
			var ticker string
			if opt1, ok := optionMap["symbol"]; ok {
				symbol = strings.ToUpper(opt1.StringValue())
				ticker = opt1.Name
				var stateMsg string
				// Adding Buy or Sell and there colored buttons
				if opt2, ok := optionMap["state"]; ok {
					state = opt2.StringValue()
					var embedColor int
					if state == "SELL" {
						maxTp = Min(tpFloats)
						stateMsg = state + " !🔴🔴🔴"
						embedColor = 0xFF4500
					} else {
						maxTp = Max(tpFloats)
						stateMsg = state + " !🔵🔵🔵"
						embedColor = 0x1E90FF
					}
					embed.URL = fmt.Sprintf("https://www.tradingview.com/chart/?symbol=%s", strings.ReplaceAll(symbol, ":", "%3A"))
					embed.Title = fmt.Sprintf("%s %s", ticker, stateMsg)
					embed.Color = embedColor
				}
			}

			// Lot size (if needed) else defualt 0.01
			var lot float32
			if opt, ok := optionMap["lot-size"]; ok {
				lot = float32(opt.FloatValue())
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "__Lot Size__",
					Value:  fmt.Sprintf("دخول ب  **%s** لكل 1,000$ من رأس المال", strconv.FormatFloat(float64(lot), 'g', -1, 32)),
					Inline: false,
				})
			} else {
				lot = 0.01
			}

			// Dumby response until real one finish
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})

			// Sending the Recommendation
			rec := AddRec(isWarn, state, ticker, sl, entryLow, entryHigh, tpString, description, lot)

			// Watting for the chart to be made
			log.Println("Watting for the chart...")
			var wg sync.WaitGroup

			wg.Add(2)

			go TelegramSendMessage(&wg, rec)

			go GetChart(&wg, state, symbol, sl, entryLow, entryHigh, lot, maxTp)

			wg.Wait()

			log.Println("Reading chart image...")

			// Reading the chart image
			filepath := filepath.Join(BaseDir, "chart", "chart-img.png")
			fileData, err := os.ReadFile(filepath)
			if err != nil {
				log.Println("Error reading file:", err)
				return
			}
			file := &discordgo.File{
				ContentType: "image/png",
				Name:        "chart-img.png",
				Reader:      bytes.NewReader(fileData),
			}

			log.Println("Chart Added!")
			// Adding the Image
			embed.Image = &discordgo.MessageEmbedImage{URL: "attachment://chart-img.png"}
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text:    "التداول قد يحمل مخاطر. يرجى التداول بحذر.",
				IconURL: "https://cdn-icons-png.flaticon.com/128/5626/5626190.png",
			}

			// Editing the Recommendation with the chart
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
				Files:  []*discordgo.File{file},
			})
		},

		// Getting latest news
		"news": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			embedMsg := &discordgo.MessageEmbed{
				Title: "Latest news",
				Color: 0x0000FF,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    "Google news",
					IconURL: "https://www.gstatic.com/gnews/logo/google_news_192.png",
				},
			}

			// Dumby response until real one finish
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embedMsg},
				},
			})

			// Getting the response
			response, err1 := AnalyisNews()
			if err1 != nil {
				log.Printf("Error fetching news: %s", err1)
			}

			// Making the embed
			embedMsg = &discordgo.MessageEmbed{
				Description: response,
			}

			// Finally sending the actual embed
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embedMsg},
			})
		},
	}
)

func SendNew(event Events) {
	file, err := os.Open("economic-calendar.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var events map[string][]Events
	if err := json.NewDecoder(file).Decode(&events); err != nil {
		log.Fatalf("Error in reading JSON! %s", err)
	}

	// Send to discord
	DgSession.ChannelMessageSend("1281269917030678713", event.Event)

	// Send to telegram
	var wg sync.WaitGroup

	wg.Add(1)

	go TelegramSendMessage(&wg, event.Event)

	wg.Wait()
}

// WARN:  not used
func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore bot message
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// Checks for messages in server
	switch {
	case strings.Contains(message.Content, "hawk"):
		discord.MessageReactionAdd(message.ChannelID, message.ID, "👀")
		discord.ChannelMessageSend(message.ChannelID, "Hey, I see you :)")
	// Rules
	case strings.ContainsFunc(message.Content, unicode.IsDigit):
		reactMsg := message.ID
		reactChannel := message.ChannelID
		res, err := CheckRules(message.Content)
		if strings.HasPrefix(res, "...") {
			return
		}
		if err != nil {
			log.Printf("Warning: %s", err)
			discord.MessageReactionRemove(reactChannel, reactMsg, discord.State.User.ID, "⚠️")
		}
		discord.ChannelMessageSendReply(message.ChannelID, res, &discordgo.MessageReference{MessageID: message.ID})
	}
}
