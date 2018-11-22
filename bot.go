package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/marcopaganini/pixiebot/reddit"
	"gopkg.in/telegram-bot-api.v4"
	"math/rand"
	//"github.com/davecgh/go-spew/spew"
)

// tgbotInterface defines an interface between this bot and the telegram API.
type tgbotInterface interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
}

// redditClientInterface defines an interface between this bot and the reddit package.
type redditClientInterface interface {
	RandomMediaURL(string) (string, int, error)
}

// run is the main message dispatcher for the bot.
func run(bot tgbotInterface, rclient redditClientInterface, triggers TriggerConfig) {
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
func handleTriggers(bot tgbotInterface, update tgbotapi.Update, rclient redditClientInterface, triggers TriggerConfig) {
	msg := update.Message.Text

	subreddit, ok, err := checkTriggers(msg, triggers)
	if err != nil {
		glog.Errorf("Error checking triggers: %v", err)
		return
	}
	if !ok {
		return
	}
	glog.Infof("Triggering fetch on %s", subreddit)

	mediaURL, mediaType, err := rclient.RandomMediaURL(subreddit)
	if err != nil {
		glog.Errorf("%v", err)
		return
	}
	switch mediaType {
	// Nothing to send
	case reddit.MediaNone:
		glog.Infof("Media URL is empty. Silently ignoring.")

	// MediaImageURL: The URL points to an image, so we can upload a
	// picture directly.
	case reddit.MediaImageURL:
		if err := sendImageURL(bot, update.Message.Chat.ID, mediaURL); err != nil {
			glog.Info(err)
		}
	// MediaFileURL: The URL points to a file (typically an MP4 file, but any
	// type playable by Telegram. In this case, we send the URL as a document.
	// upload.
	case reddit.MediaFileURL:
		if err := sendFileURL(bot, update.Message.Chat.ID, mediaURL); err != nil {
			glog.Info(err)
		}
	// Video URL: Simple video url, like youtube. Telegram takes charge of
	// reading the link and generating a thumbnail.
	case reddit.MediaVideoURL:
		if err := sendURL(bot, update.Message.Chat.ID, mediaURL); err != nil {
			glog.Info(err)
		}
	}
	return
}

// sendImageURL sends a photo pointed to by mediaURL to the telegram chat
// identified by chatID using NewPhotoUpload. This is the ideal way to
// send URLs that point directly to images, which will immediately show
// in the group.
func sendImageURL(bot tgbotInterface, chatID int64, mediaURL string) error {
	// Issue #74 is at play here, preventing us to upload via url.URL:
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/74
	img := tgbotapi.NewPhotoUpload(chatID, nil)
	img.FileID = mediaURL
	img.UseExisting = true

	glog.Infof("Sending Image URL: %v\n", img)
	_, err := bot.Send(img)
	if err != nil {
		return fmt.Errorf("error sending photo (url: %s): %v", mediaURL, err)
	}
	return nil
}

// sendURL sends the media URL as a regular message to the user/group.
func sendURL(bot tgbotInterface, chatID int64, mediaURL string) error {
	msg := tgbotapi.NewMessage(chatID, mediaURL)

	glog.Infof("Sending URL: %v\n", msg)
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending media URL (url: %s): %v", mediaURL, err)
	}

	return nil
}

// sendFileURL sends the media URL that points to a Telegram playable file
// (usually an MP4 video) using NewDocumentUpload. Use sendPhoto instead if
// the URL points directly to an image.
func sendFileURL(bot tgbotInterface, chatID int64, mediaURL string) error {
	doc := tgbotapi.NewDocumentUpload(chatID, nil)
	doc.FileID = mediaURL
	doc.UseExisting = true

	glog.Infof("Sending File URL: %v\n", doc)
	_, err := bot.Send(doc)
	if err != nil {
		return fmt.Errorf("error sending file URL (url: %s): %v", mediaURL, err)
	}

	return nil
}

// checkTriggers returns the name of a subreddit if the current message matches any of the
// trigger messages configured for that subreddit.
func checkTriggers(msg string, triggers TriggerConfig) (string, bool, error) {
	for _, rule := range triggers {
		// Attempt to match regexp.
		if !rule.regex.MatchString(msg) {
			continue
		}
		// Throw dice on percentage.
		rnd := (rand.Int() % 100) + 1
		if rule.percentage <= rnd {
			glog.Infof("No dice for subreddit %s! Wanted [1-%d], got %d\n", rule.subreddit, rule.percentage, rnd)
			continue
		}
		return rule.subreddit, true, nil
	}
	return "", false, nil
}
