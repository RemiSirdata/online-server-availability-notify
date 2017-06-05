package main

import (
	"github.com/nlopes/slack"
	"fmt"
	"time"
	"strconv"
	"strings"
)

type Bot struct {
	AppContext     *AppContext
	Crawler        *Crawler
	StartTimestamp time.Time
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
	b.StartTimestamp = time.Now()
	b.updateChannel()
	b.ListenMembers()
}

func (b *Bot) ListenMembers() {
	rtm := b.GetClient().NewRTM()
	go rtm.ManageConnection()
	botIdentity, err := b.GetClient().GetUserIdentity()
	botId := ""
	if err != nil {
		b.AppContext.Logger.Error(fmt.Sprintf("Fail to get connected user info"))
	} else {
		botId = botIdentity.User.ID
	}

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
			if ev.User == botId {
				b.AppContext.Logger.Debug("Message written by bot")
				break
			}
			userInfo, err := rtm.GetUserInfo(ev.User)
			if err != nil {
				b.AppContext.Logger.Error(fmt.Sprintf("Fail to get user info for %s", ev.User))
				break
			}
			b.parseUserMessage(ev, userInfo)
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

func (b *Bot) parseUserMessage(event *slack.MessageEvent, userInfo *slack.User) {
	t, err := strconv.ParseFloat(event.Timestamp, 64)
	if err != nil {
		b.AppContext.Logger.Error(fmt.Sprintf("Fail to parse timestamp %s", event.Timestamp))
		return
	}
	if b.StartTimestamp.Unix() > int64(t) {
		b.AppContext.Logger.Debug("message written before start, ignore")
		return
	}
	text := strings.TrimSpace(event.Text)
	switch b.getCommandType(text) {
	case COMMAND_LIST_SERVER:
		b.DumpServerList()
		break
	case COMMAND_SUBSCRIBE_SERVER_UPDATE:
		b.Subscribe(text, event, userInfo)
	}
}

func (b *Bot) getCommandType(text string) string {
	if strings.Contains(text, " ") {
		explode := strings.Split(text, " ")
		return explode[0]
	}
	return strings.ToLower(text)
}

func (b *Bot) DumpServerList() {
	for _, server := range b.Crawler.ServerList.Servers {
		b.GetClient().PostMessage(b.AppContext.ChannelName, server.GetStatus(), slack.PostMessageParameters{})
	}
}

func (b *Bot) Subscribe(text string, event *slack.MessageEvent, userInfo *slack.User) {
	serverName := strings.TrimSpace(text[len(COMMAND_SUBSCRIBE_SERVER_UPDATE):])
	server, found := b.Crawler.ServerList.GetServerByName(serverName)
	message := fmt.Sprintf(MESSAGE_SERVER_NOT_FOUND, serverName)
	if found {
		message = fmt.Sprintf(MESSAGE_SUBSCRIBE_SERVER, server.Name)
		go func() {
			listener := b.Crawler.AddListenerForServer(serverName)
			for server := range listener.Chan {
				message := fmt.Sprintf(MESSAGE_NOTIFY_SERVER_UPDATE, userInfo.Name, serverName, server.ServerAvailable, server.PreviousAvailability)
				b.NotifyUser(userInfo, message)
			}
		}()
	}
	b.GetClient().PostMessage(b.AppContext.ChannelName, message, slack.PostMessageParameters{})
}

func (b *Bot) NotifyUser(userInfo *slack.User, message string) {
	b.GetClient().PostMessage(b.AppContext.ChannelName,message, slack.PostMessageParameters{})
}

func (b *Bot) updateChannel() {
	go func() {
		listener := b.Crawler.AddListener()
		for server := range listener.Chan {
			if server.ServerAvailable != server.PreviousAvailability {
				b.AppContext.Logger.Debug(fmt.Sprintf("Server update : %v", server))
				_, t, err := b.GetClient().PostMessage(b.AppContext.ChannelName, server.GetAvailabilityMessage(), slack.PostMessageParameters{})
				if err != nil {
					b.AppContext.Logger.Error("Fail to add Message to channel %s", b.AppContext.ChannelName)
				} else {
					b.AppContext.Logger.Debug(fmt.Sprintf("Message successfully added to channel (%s) %s : %s", t, b.AppContext.ChannelName, server.GetAvailabilityMessage()))
				}
			}
		}
	}()
}
