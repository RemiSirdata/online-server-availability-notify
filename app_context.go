package main

import (
	"github.com/nlopes/slack"
	"github.com/inconshreveable/log15"
)

type AppContext struct {
	Client      *slack.Client
	Channel     *slack.Channel
	ChannelName string
	Logger      log15.Logger
}
