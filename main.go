package main

import (
	"sync"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"uav-bot/bot/bot"
)


func init() {
	// Getting curret file
	bot.BaseDir = "."
	log.Printf("Base directory is: %s", bot.BaseDir)

	var err error
	dtoken, _ := os.LookupEnv("DISCORD_TOKEN")
	bot.DgSession, err = discordgo.New("Bot " + dtoken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	bot.DgSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := bot.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	bot.DgSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Hawk Flying...")
	})
	err := bot.DgSession.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(bot.Commands))
	for i, v := range bot.Commands {
		cmd, err := bot.DgSession.ApplicationCommandCreate(bot.DgSession.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Schedule events immediately at startup
	if time.Now().UTC().Weekday() == time.Sunday {
		var wg sync.WaitGroup

		wg.Add(1)

		go bot.InvestNews(&wg)

		wg.Wait()
	} else {
		go bot.ScheduleEvents()
	}

	// Start the gocron scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	// Schedule the daily task
	scheduler.Every(1).Day().At("20:00").Do(bot.ScheduleEvents)

	// Schedule the weekly task
	scheduler.Every(1).Sunday().At("00:00").Do(func() {
		var wg sync.WaitGroup

		wg.Add(1)

		go bot.InvestNews(&wg)

		wg.Wait()
	})

	scheduler.StartAsync()

	defer bot.DgSession.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Removing commands...")

	for _, v := range registeredCommands {
		err := bot.DgSession.ApplicationCommandDelete(bot.DgSession.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		} else {
			log.Printf("Deleted command: %v", v.Name)
		}
	}

	log.Println("Hawk landed!")
}
