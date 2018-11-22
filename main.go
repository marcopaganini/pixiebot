package main

import (
	"github.com/marcopaganini/pixiebot/reddit"
	"gopkg.in/telegram-bot-api.v4"
	"log"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// New Reddit client.
	rclient := reddit.NewClient(config.Username, config.Password, config.ClientID, config.Secret)

	// New Bot.
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}

	// run bot (this should never exit).
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	run(bot, rclient, config.triggerConfig)
}
