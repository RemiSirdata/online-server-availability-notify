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
	b.ListenMembers()
}

func (b *Bot) ListenMembers() {
	rtm := b.GetClient().NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		b.AppContext.Logger.Debug(fmt.Sprintf("Event data %s", msg))
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Infos: %v", ev.Info))
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Connection counter %d", ev.ConnectionCount))
			break
		case *slack.ConnectingEvent:
			break

		case *slack.MessageEvent:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Message from %s: %v\n", ev.User, ev))
			_, err := rtm.GetUserInfo(ev.User)
			if err != nil {
				b.AppContext.Logger.Error(fmt.Sprintf("Fail to get user info for %s", ev.User))
				return
			}
			break

		case *slack.PresenceChangeEvent:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Presence Change: %v\n", ev))
			break

		case *slack.LatencyReport:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Current latency: %v\n", ev.Value))
			break

		case *slack.RTMError:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Error: %s\n", ev.Error()))
			break

		case *slack.InvalidAuthEvent:
			b.AppContext.Logger.Debug("[Event] Invalid credentials")
			break

		default:
			b.AppContext.Logger.Debug(fmt.Sprintf("[Event] Unidentified %s", msg.Type))
		}
	}

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
