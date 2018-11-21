package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/marcopaganini/pixiebot/reddit"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	flag.Parse() // glog needs this.
	defer glog.Flush()

	config, err := loadConfig()
	if err != nil {
		glog.Exit(err)
	}

	// New Reddit client.
	rclient := reddit.NewClient(config.Username, config.Password, config.ClientID, config.Secret)

	// New Bot.
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		glog.Exitf("Error starting bot: %v", err)
	}

	// run bot (this should never exit).
	bot.Debug = true
	glog.Infof("Authorized on account %s", bot.Self.UserName)
	run(bot, rclient, config.triggerConfig)
}
