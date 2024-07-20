package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/xgnid-tw/gkmas-last-cal-bot/discord"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("can not fetch env")
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	devAPPID := os.Getenv("DISCORD_APP_ID")

	dc, err := discordgo.New(discordToken)
	if err != nil {
		log.Fatalf("can not create discord session, %s", err)
	}

	des := discord.NewDiscordEventService(dc, devAPPID)

	dc.AddHandler(des.InteractCommand)

	dc.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	dc.AddHandler(des.RegisterSlashCommand)

	err = dc.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = dc.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
