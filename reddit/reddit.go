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

// RandomPicURL returns the URL containing a random picture from a given
// subreddit.  Returns the empty string and nil (no error) if random article
// does not contain any pictures.
func (c *Client) RandomPicURL(subreddit string) (string, error) {
	// Refresh token, if needed.
	if err := c.cred.RefreshToken(); err != nil {
		return "", err
	}

	// Create request to the OAuth enabled URL with all tokens.
	redditURL := fmt.Sprintf(c.randomArticleURL, subreddit)

	req, err := http.NewRequest("GET", redditURL, nil)
	if err != nil {
		return "", err
	}
	tok, err := c.cred.Token()
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("error fetching reddit URL: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reddit returned code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Decode JSON looking for the "image" attribute.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Random response is structured as follows:
	// - [0] (type listing): contains the original message.
	// - data: contains all children.
	// - children: (type Listing): contains the multiple image formats.
	// - [1:n] (type listing): contains children with the comments.
	//
	// We parse the message twice: Once to obtain "data", which contains all the
	// children with the image previews. Data should exist in all cases, so we
	// return an error if we get one here. We then parse data itself to obtain the
	// image URLs. Not all responses have image URLs, so we return "" if an error
	// happens here.
	rdata, _, _, err := jsonparser.Get(body, "[0]", "data", "children", "[0]", "data")
	if err != nil {
		glog.Errorf("Error decoding 'data' in json: %v, body: %s", err, body)
		return "", err
	}

	val, _, _, err := jsonparser.Get(rdata, "preview", "images", "[0]", "source", "url")
	if err != nil {
		glog.Infof("Can't find 'preview' in json: %v, body: %s", err, body)
		return "", nil
	}
	imgURL := html.UnescapeString(string(val))
	glog.Infof("Image URL: %s", imgURL)
	return imgURL, nil
}
