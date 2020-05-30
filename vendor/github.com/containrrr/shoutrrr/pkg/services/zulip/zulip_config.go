package zulip

import (
	"errors"
	"net/url"
)

// Config for the zulip service
type Config struct {
	BotMail string
	BotKey  string
	Host    string
	Path    string
	Stream  string
	Topic   string
}

// GetURL returns a URL representation of it's current field values
func (config *Config) GetURL() *url.URL {
	return &url.URL{
		User:   url.UserPassword(config.BotMail, config.BotKey),
		Host:   config.Host,
		Path:   config.Path,
		Scheme: Scheme,
	}
}

// SetURL updates a ServiceConfig from a URL representation of it's field values
func (config *Config) SetURL(serviceURL *url.URL) error {
	var ok bool

	config.BotMail = serviceURL.User.Username()

	if config.BotMail == "" {
		return errors.New(string(MissingBotMail))
	}

	config.BotKey, ok = serviceURL.User.Password()

	if !ok {
		return errors.New(string(MissingAPIKey))
	}

	config.Host = serviceURL.Hostname()

	if config.Host == "" {
		return errors.New(string(MissingHost))
	}

	config.Path = "api/v1/messages"
	config.Stream = serviceURL.Query().Get("stream")
	config.Topic = serviceURL.Query().Get("topic")

	return nil
}

const (
	// Scheme is the identifying part of this service's configuration URL
	Scheme = "zulip"
)

// CreateConfigFromURL to use within the zulip service
func CreateConfigFromURL(serviceURL *url.URL) (*Config, error) {
	config := Config{}
	err := config.SetURL(serviceURL)

	return &config, err
}
