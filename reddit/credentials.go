package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// redditAuthURL contains the URL for reddit authorization
	// (exchanging user/password for access-token).
	redditAuthURL = "https://www.reddit.com/api/v1/access_token"
)

// Token holds the Oauth2 token from Reddit.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`

	// Token creation time.
	ctime time.Time
}

// Credentials holds all state require to authenticate a reddit request.
type Credentials struct {
	token        *Token
	username     string
	password     string
	clientID     string
	clientSecret string

	// Reddit auth URL.
	tokenURL string
}

// NewCredentials returns a pointer to a new Credentials object.
func NewCredentials(username, password, clientID, clientSecret string) *Credentials {
	return &Credentials{
		username:     username,
		password:     password,
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     redditAuthURL,
	}
}

// RefreshToken fetches a new authorization token (if needed).
func (c *Credentials) RefreshToken() error {
	if c == nil {
		return errors.New("unitialized credentials")
	}

	// Do we need a new token?
	if validToken(c.token) {
		return nil
	}

	client := &http.Client{}

	// To create a new "Reddit App", visit https://www.reddit.com/prefs/apps
	// username: reddit username.
	// password: reddit password.
	v := url.Values{}
	v.Set("grant_type", "password")
	v.Set("username", c.username)
	v.Set("password", c.password)

	formdata := v.Encode()

	var try uint
	var resp *http.Response

	for try = 0; try < authTries; try++ {
		req, err := http.NewRequest("POST", redditAuthURL, strings.NewReader(formdata))
		if err != nil {
			return fmt.Errorf("error creating HTTP request: %v", err)
		}
		req.SetBasicAuth(c.clientID, c.clientSecret)
		req.Header.Add("User-agent", userAgent)
		resp, err = client.Do(req)

		if err != nil {
			return fmt.Errorf("token request error: %v", err)
		}
		defer resp.Body.Close()

		// All good?
		if resp.StatusCode == http.StatusOK {
			break
		}

		// Overloaded?
		if resp.StatusCode == http.StatusTooManyRequests {
			d := time.Duration(1<<try) * time.Second
			log.Printf("Server busy. Will retry in %v", d)
			time.Sleep(d)
			continue
		}

		// Anything else == error
		return fmt.Errorf("token request error: http %v", resp.StatusCode)
	}

	if try == authTries {
		return fmt.Errorf("unable to authenticate (server overloaded)")
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read token request response: %v", err)
	}
	if err := json.Unmarshal(buf, &c.token); err != nil {
		return fmt.Errorf("unable to decode reddit auth token: %v", err)
	}
	// Set token last updated time.
	c.token.ctime = time.Now()

	return nil
}

// Token returns the latest token (or triggers a token fetch, if needed).
func (c *Credentials) Token() (*Token, error) {
	if err := c.RefreshToken(); err != nil {
		return nil, err
	}
	return c.token, nil
}

// validToken returns true if token is still valid. False otherwise.
func validToken(token *Token) bool {
	// Non-initialized token == invalid token.
	if token == nil {
		return false
	}

	// New expiration time. We remove 30 seconds to give the caller
	// some time with a valid token.
	exp := token.ctime.Add(time.Duration(token.ExpiresIn) * time.Second)
	exp = exp.Add(-30 * time.Second)

	now := time.Now()
	if now.Before(exp) {
		log.Printf("Token still valid: now: %v, expires: %v", time.Now(), exp)
		return true
	}
	return false
}
