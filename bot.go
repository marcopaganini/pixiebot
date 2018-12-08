package main

import (
	"fmt"
	"github.com/marcopaganini/pixiebot/reddit"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"math/rand"
	"time"
	//"github.com/davecgh/go-spew/spew"
)

const (
	// Default sleep period.
	sleepTime = time.Hour

	// Default time format.
	timeFormat = "2006-01-02 15:04:05 MST"
)

type tgbotSender interface {
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
}

// redditClientInterface defines an interface between this bot and the reddit package.
type redditClientInterface interface {
	RandomMediaURL(string) (string, int, error)
}

// botSleepTime keeps the time of the last request for the bot to sleep, per group.
type botSleepTime map[int64]time.Time

// run is the main message dispatcher for the bot.
func run(bot tgbotSender, updates tgbotapi.UpdatesChannel, rclient redditClientInterface, triggers TriggerConfig) {
	bsleep := botSleepTime{}

	for update := range updates {
		if update.Message == nil || update.Message.From.IsBot {
			continue
		}

		chatID := update.Message.Chat.ID

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(chatID, "")

			switch update.Message.Command() {
			case "sleep":
				wake := time.Now().Add(sleepTime)
				bsleep[chatID] = wake
				msg.Text = fmt.Sprintf("Sleeping until %s. Zzzzz...", wake.Format(timeFormat))
			case "wakeup":
				bsleep[chatID] = time.Now().Add(time.Minute * -1)
				msg.Text = "Fully awake and ready to serve!"
			default:
				continue
			}
			bot.Send(msg)
			continue
		}

		if sleeping(bsleep, chatID) {
			continue
		}

		handleTriggers(bot, update, rclient, triggers)
	}
}

// sleeping returns true if the bot is still sleeping, false otherwise.
func sleeping(bsleep botSleepTime, id int64) bool {
	if t, ok := bsleep[id]; ok && time.Now().Before(t) {
		return true
	}
	return false
}

// handleTriggers checks if the message is a trigger message and emits a picture
// from the configured subreddit if so.
func handleTriggers(bot tgbotSender, update tgbotapi.Update, rclient redditClientInterface, triggers TriggerConfig) {
	handlers := map[int]func(tgbotSender, int64, string) error{
		// MediaNone: Nothing to do...
		reddit.MediaNone: nil,

		// MediaImageURL: The URL points to an image, so we can upload a
		// picture directly.
		reddit.MediaImageURL: sendImageURL,

		// MediaFileURL: The URL points to a file (typically an MP4 file, but
		// any type playable by Telegram. In this case, we send the URL as a
		// document.  upload.
		reddit.MediaFileURL: sendFileURL,

		// Video URL: Simple video url, like youtube. Telegram takes charge of
		// reading the link and generating a thumbnail.
		reddit.MediaVideoURL: sendURL,
	}

	msg := update.Message.Text

	subreddit, ok, err := checkTriggers(msg, triggers)
	if err != nil {
		log.Printf("Error checking triggers: %v", err)
		return
	}
	if !ok {
		return
	}
	log.Printf("Triggering fetch on %s", subreddit)

	// Dispatch handler using mediaType as key in handlers.
	mediaURL, mediaType, err := rclient.RandomMediaURL(subreddit)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	handler, ok := handlers[mediaType]
	if !ok {
		log.Printf("Media URL is empty. Silently ignoring.")
		return
	}

	if err := handler(bot, update.Message.Chat.ID, mediaURL); err != nil {
		log.Print(err)
	}
}

// sendImageURL sends a photo pointed to by mediaURL to the telegram chat
// identified by chatID using NewPhotoUpload. This is the ideal way to
// send URLs that point directly to images, which will immediately show
// in the group.
func sendImageURL(bot tgbotSender, chatID int64, mediaURL string) error {
	// Issue #74 is at play here, preventing us to upload via url.URL:
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/74
	img := tgbotapi.NewPhotoUpload(chatID, nil)
	img.FileID = mediaURL
	img.UseExisting = true

	log.Printf("Sending Image URL: %v\n", img)
	_, err := bot.Send(img)
	if err != nil {
		return fmt.Errorf("error sending photo (url: %s): %v", mediaURL, err)
	}
	return nil
}

// sendURL sends the media URL as a regular message to the user/group.
func sendURL(bot tgbotSender, chatID int64, mediaURL string) error {
	msg := tgbotapi.NewMessage(chatID, mediaURL)

	log.Printf("Sending URL: %v\n", msg)
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending media URL (url: %s): %v", mediaURL, err)
	}

	return nil
}

// sendFileURL sends the media URL that points to a Telegram playable file
// (usually an MP4 video) using NewDocumentUpload. Use sendPhoto instead if
// the URL points directly to an image.
func sendFileURL(bot tgbotSender, chatID int64, mediaURL string) error {
	doc := tgbotapi.NewDocumentUpload(chatID, nil)
	doc.FileID = mediaURL
	doc.UseExisting = true

	log.Printf("Sending File URL: %v\n", doc)
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
			log.Printf("No dice for subreddit %s! Wanted [1-%d], got %d\n", rule.subreddit, rule.percentage, rnd)
			continue
		}
		return rule.subreddit, true, nil
	}
	return "", false, nil
}
