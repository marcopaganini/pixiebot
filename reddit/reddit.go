package reddit

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// redditRandomArticleURL contains the default format for a reddit URL
	// used to fetch a random article from a subreddit.
	kRandomArticleURL = "http://oauth.reddit.com/r/%s/random.json"

	// Custom user-agent is recommended by reddit.
	// Most default UAs are heavily throttled.
	userAgent = "github.com/marcopaganini/pixiebot"

	// authTries defines how many times authentication is
	// attempted before giving up.
	authTries = 5
)

// Media types
const (
	MediaNone     = iota // 0: No usable media.
	MediaImageURL = iota // 1: An URL pointing to an image.
	MediaFileURL  = iota // 2: An URL pointing to a file.
	MediaVideoURL = iota // 3: An URL pointing to a video.
)

// Client holds state about a Reddit client
type Client struct {
	cred CredentialsInterface

	// Format of the random article URL (expands subreddit).
	randomArticleURL string
}

// CredentialsInterface defines the interface between the client and
// the Credentials routines.
type CredentialsInterface interface {
	RefreshToken() error
	Token() (*Token, error)
}

// NewClient creates a new Reddit client using the passed credentials.
func NewClient(username, password, clientID, clientSecret string) *Client {
	return &Client{
		cred:             NewCredentials(username, password, clientID, clientSecret),
		randomArticleURL: kRandomArticleURL,
	}
}

// RandomMediaURL returns the URL containing a random media from a given
// subreddit. The type specifies the type of media being returned (usually an
// URL pointing to an image or to a video). Returns the type empty string with
// type mediaNone if the random article does not contain any pictures.
func (c *Client) RandomMediaURL(subreddit string) (string, int, error) {
	// Refresh token, if needed.
	if err := c.cred.RefreshToken(); err != nil {
		return "", MediaNone, err
	}

	// Create request to the OAuth enabled URL with all tokens.
	redditURL := fmt.Sprintf(c.randomArticleURL, subreddit)

	req, err := http.NewRequest("GET", redditURL, nil)
	if err != nil {
		return "", MediaNone, err
	}
	tok, err := c.cred.Token()
	if err != nil {
		return "", MediaNone, err
	}
	req.Header.Add("Authorization", "bearer "+tok.AccessToken)
	req.Header.Add("User-agent", userAgent)

	// Create an http client that forwards all headers in case of redirection.
	// (default behavior for Go http is to not forward Auth to other domains.)
	httpClient := &http.Client{
		CheckRedirect: func(redir *http.Request, via []*http.Request) error {
			// Add an authorization: bearer <token> header if the destination
			// contains the substring "oauth" (otherwise, we don't need it.)
			var loc *url.URL
			loc, err = redir.Response.Location()
			if err != nil {
				return err
			}
			if strings.Contains(loc.Hostname(), "oauth") {
				redir.Header.Add("Authorization", "bearer "+tok.AccessToken)
			}

			// Allow 10 redirects.
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", MediaNone, fmt.Errorf("error fetching reddit URL: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", MediaNone, fmt.Errorf("reddit returned code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Decode JSON looking for the "image" attribute.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", MediaNone, err
	}

	return media(body)
}

// media returns the type of media and media URL for a given json 'list' type.
//
// It assumes a few things about the JSON input:
// - [0] (type listing): contains the original message.
// - data: contains all children.
// - children: (type Listing): contains the multiple image formats.
// - [1:n] (type listing): contains children with the comments.
//
func media(data []byte) (string, int, error) {
	// We parse the message twice: Once to obtain "data", which contains all
	// the children with the information we need. Data should exist in all
	// cases, so we return an error if we get one here. We then parse data
	// itself to obtain the media or image preview URLs. Not all responses have
	// image URLs, so we return "" if an error happens here.
	rdata, _, _, err := jsonparser.Get(data, "[0]", "data", "children", "[0]", "data")
	if err != nil {
		return "", MediaNone, fmt.Errorf("Error decoding 'data' in json: %v: %v", data, err)
	}

	// Look for data.media.type. Parsing errors mean we don't have a "type"
	// field, so they don't really matter (we just log). If for youtube and
	// gfycat, data.url has the URL of the post.
	dtype, err := jsonparser.GetString(rdata, "media", "type")
	if err == nil {
		if dtype == "youtube.com" || dtype == "gfycat.com" {
			glog.Infof("Returning media type: %s", dtype)
			u, err := jsonparser.GetString(rdata, "url")
			return html.UnescapeString(u), MediaVideoURL, err
		}
	}

	// Reddit media posts contain a data.reddit_video entry.  In this case,
	// data.url points to the article URL and data.reddit_video.fallback_url
	// points to the data. We need to serve as a file (which signals to the
	// client to use NewDocument when posting this link).
	u, err := jsonparser.GetString(rdata, "media", "reddit_video", "fallback_url")
	if err == nil {
		glog.Infof("Returning a data.media.fallback_url (reddit_video)")
		return html.UnescapeString(u), MediaFileURL, nil
	}

	// Posts with a data.url ending in gif or gifv.
	u, err = jsonparser.GetString(rdata, "url")
	if strings.HasSuffix(u, "gif") || strings.HasSuffix(u, "gifv") {
		glog.Infof("Returning GIF/GIFv URL")
		return html.UnescapeString(u), MediaImageURL, nil
	}

	// At this point, we check for regular preview images.
	u, err = jsonparser.GetString(rdata, "preview", "images", "[0]", "source", "url")
	if err != nil {
		glog.Infof("Can't find 'preview' in json: %v", err)
		return "", MediaNone, nil
	}
	imgURL := html.UnescapeString(u)
	glog.Infof("Plain image preview URL: %s", imgURL)
	return imgURL, MediaImageURL, nil
}
