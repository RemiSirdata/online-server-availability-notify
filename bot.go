package main

import (
	"github.com/nlopes/slack"
	"fmt"
)

type Bot struct {
	AppContext *AppContext
	Crawler    *Crawler
}

func newBot(crawler *Crawler, context *AppContext) *Bot {
	return &Bot{
		Crawler:    crawler,
		AppContext: context,
	}
}

func (b *Bot) GetClient() *slack.Client {
	return b.AppContext.Client
}

func (b *Bot) Start() {
	b.updateChannel()
}

func (b *Bot) updateChannel() {
	go func() {
		listener := b.Crawler.AddListener()
		for server := range listener.Chan {
			b.AppContext.Logger.Debug(fmt.Sprintf("Server update : %v", server))
			_, t, err := b.GetClient().PostMessage(b.AppContext.ChannelName, server.GetAvailabilityMessage(), slack.PostMessageParameters{})
			if err != nil {
				b.AppContext.Logger.Error("Fail to add Message to channel %s", b.AppContext.ChannelName)
			} else {
				b.AppContext.Logger.Debug(fmt.Sprintf("Message successfully added to channel (%s) %s : %s", t, b.AppContext.ChannelName, server.GetAvailabilityMessage()))
			}
		}
	}()
}
