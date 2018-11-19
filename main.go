package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/marcopaganini/pixiebot/reddit"
	"gopkg.in/telegram-bot-api.v4"
	"strings"
)

// tgbotInterface defines an interface between this bot and the telegram API.
type tgbotInterface interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
}

// redditClientInterface defines an interface between this bot and the reddit package.
type redditClientInterface interface {
	RandomPicURL(string) (string, error)
}

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
	run(bot, rclient, config.Triggers)
}

// run is the main message dispatcher for the bot.
func run(bot tgbotInterface, rclient redditClientInterface, triggers ConfigTriggers) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)
	for update := range updates {
		// Check trigger messages.
		if update.Message != nil {
			handleTriggers(bot, update, rclient, triggers)
		}
	}
}

// handleTriggers checks if the message is a trigger message and emits a picture
// from the configured subreddit if so.
func handleTriggers(bot tgbotInterface, update tgbotapi.Update, rclient redditClientInterface, triggers ConfigTriggers) {
	msg := strings.ToLower(update.Message.Text)
	glog.Infof("Checking %q", msg)
	subreddit, _, ok := checkTriggers(msg, triggers)
	if !ok {
		return
	}
	glog.Infof("Triggering fetch on %s", subreddit)

	// Get a random picture URL and download into a temporary file.
	mediaURL, err := rclient.RandomPicURL(subreddit)
	if err != nil {
		glog.Errorf("%v", err)
		return
	}
	if mediaURL == "" {
		glog.Infof("Media URL is empty. Silently ignoring.")
		return
	}
	if err := sendPhoto(bot, update.Message.Chat.ID, mediaURL); err != nil {
		glog.Info(err)
	}

	return
}

// sendPhoto sends a photo pointed to by mediaURL to the telegram chat identified by chatID.
func sendPhoto(bot tgbotInterface, chatID int64, mediaURL string) error {
	// Issue #74 is at play here, preventing us to upload via url.URL:
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/74
	photoMsg := tgbotapi.NewPhotoUpload(chatID, nil)
	photoMsg.FileID = mediaURL
	photoMsg.UseExisting = true

	glog.Infof("Sending %v\n", photoMsg)
	_, err := bot.Send(photoMsg)
	if err != nil {
		return fmt.Errorf("error sending photo (url: %s): %v", mediaURL, err)
	}
	return nil
}

// checkTriggers returns the name of a subreddit if the current message matches any of the
// trigger messages configured for that subreddit.
func checkTriggers(msg string, triggers ConfigTriggers) (string, ConfigTrigger, bool) {
	for subreddit, trigger := range triggers {
		for _, w := range trigger.Keywords {
			if strings.Contains(msg, w) {
				return subreddit, trigger, true
			}
		}
	}
	return "", ConfigTrigger{}, false
}
