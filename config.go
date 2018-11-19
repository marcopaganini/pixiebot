package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const (
	configFile = "config.toml"

	// Directory usually under $HOME/.config that holds all configurations.
	botConfigDir = "pixiebot"
)

type ConfigTrigger struct {
	Subreddit  string   `toml:"subreddit"`
	Keywords   []string `toml:"keywords"`
	Percentage int      `toml:"percentage"`
	WaitSec    int      `toml:"waitsec"`
}

type ConfigTriggers map[string]ConfigTrigger

type botConfig struct {
	// Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	ClientID string `toml:"client_id"`
	Secret   string `toml:"secret"`

	// Telegram Token
	Token string `toml:"token"`

	// Triggers with subreddit as the key.
	Triggers ConfigTriggers `toml:"triggers"`
}

// loadConfig loads the configuration items for the bot from 'configFile' under
// the home directory, and assigns sane defaults to certain configuration items.
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

// dataDir returns the location for data files. Use the XDG_DATA_HOME
// environment variable, or the fallback value of $HOME/.local/share if the variable
// is not set. It also attempts to create dataDir, in case it does not exist.
func dataDir() (string, error) {
	xdg := os.Getenv("XDG_DATA_HOME")

	if xdg != "" {
		dir := filepath.Join(xdg, botConfigDir)
		return dir, os.MkdirAll(dir, 0755)
	}

	home, err := homeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", botConfigDir)
	return dir, os.MkdirAll(dir, 0755)
}
