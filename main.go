package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"

	"uav-bot/bot/bot"
)

var s *discordgo.Session

func init() {
	// Getting curret file
	bot.BaseDir = "."
	log.Printf("Base directory is: %s", bot.BaseDir)

	var err error
	dtoken, _ := os.LookupEnv("DISCORD_TOKEN")
	s, err = discordgo.New("Bot " + dtoken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := bot.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Hawk Flying...")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(bot.Commands))
	for i, v := range bot.Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Removing commands...")

	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		} else {
			log.Printf("Deleted command: %v", v.Name)
		}
	}

	log.Println("Hawk landed!")
}
