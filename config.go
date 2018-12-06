package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
)

const (
	configFile = "config.toml"

	// Directory usually under $HOME/.config that holds all configurations.
	botConfigDir = "pixiebot"
)

// TOMLTriggerRule represents a configuration map in TOML.
type TOMLTriggerRule struct {
	Subreddit  string `toml:"subreddit"`
	Regex      string `toml:"regex"`
	Percentage int    `toml:"percentage"`
}

// TOMLTriggerConfig is a map of TOML trigger configs.
type TOMLTriggerConfig map[string]TOMLTriggerRule

// TriggerRule stores the in-memory (parsed & sanitized) trigger config.
type TriggerRule struct {
	subreddit  string
	regex      *regexp.Regexp
	percentage int
}

// TriggerConfig holds a collection of trigger rules.
type TriggerConfig []TriggerRule

// botConfig stores configuration about this bot instance.
type botConfig struct {
	// Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	ClientID string `toml:"client_id"`
	Secret   string `toml:"secret"`

	// Telegram Token
	Token string `toml:"token"`

	// Trigger config as represented in the TOML file.
	TOMLTriggerConfig TOMLTriggerConfig `toml:"triggers"`

	// Parsed and sanitized trigger config.
	triggerConfig TriggerConfig
}

// loadConfig loads the configuration items for the bot from 'configFile' under
// the home directory, and assigns sane defaults to certain configuration
// items.  Returns a filled-in botConfig object.
func loadConfig() (botConfig, error) {
	config := botConfig{}

	cfgdir, err := configDir()
	if err != nil {
		return botConfig{}, err
	}
	f := filepath.Join(cfgdir, configFile)

	buf, err := ioutil.ReadFile(f)
	if err != nil {
		return botConfig{}, err
	}
	if _, err := toml.Decode(string(buf), &config); err != nil {
		return botConfig{}, err
	}

	// Check mandatory fields.
	if config.Username == "" || config.Password == "" || config.ClientID == "" || config.Secret == "" {
		return botConfig{}, errors.New("usename/password/client_id/secret cannot be null")
	}

	// Generate triggerConfig.
	// First, fetch every key and order. We want keys to be processed in sequence.
	var keys []string

	for k := range config.TOMLTriggerConfig {
		keys = append(keys, k)
	}

	// Now add multiple trigger configs in key order.
	tc := TriggerConfig{}
	for _, k := range keys {
		fileRule := config.TOMLTriggerConfig[k]

		// Check percentage.
		if fileRule.Percentage < 0 || fileRule.Percentage > 100 {
			return botConfig{}, fmt.Errorf("trigger percentage must be between 0 and 100, got %d", fileRule.Percentage)
		}

		tr := TriggerRule{}
		tr.subreddit = fileRule.Subreddit
		tr.percentage = fileRule.Percentage

		// Convert regex to a compiled object for later use.
		var err error
		tr.regex, err = regexp.Compile(fileRule.Regex)
		if err != nil {
			return botConfig{}, fmt.Errorf("rule contains invalid regex: %q: %v", fileRule.Regex, err)
		}
		tc = append(tc, tr)
	}
	config.triggerConfig = tc

	return config, nil
}

// homeDir returns the user's home directory or an error if the variable HOME
// is not set, or os.user fails, or the directory cannot be found.
func homeDir() (string, error) {
	// Get home directory from the HOME environment variable first.
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("reading user info: %v", err)
		}
		home = usr.HomeDir
	}
	_, err := os.Stat(home)
	if os.IsNotExist(err) || !os.ModeType.IsDir() {
		return "", fmt.Errorf("homedir must exist: %s", home)
	}
	// Other errors than file not found.
	if err != nil {
		return "", err
	}
	return home, nil
}

// configDir returns the location for config files. Use the XDG_CONFIG_HOME
// environment variable, or the fallback value of $HOME/.config if the variable
// is not set.
func configDir() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, botConfigDir), nil
	}
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", botConfigDir), nil
}
