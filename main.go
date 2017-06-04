package main

import (
	"github.com/nlopes/slack"
	"flag"
	"github.com/inconshreveable/log15"
	"os"
	"time"
)

func main() {

	token := flag.String("token", "", "Slack token API")
	channelName := flag.String("channelName", DEFAULT_CHANNEL_NAME, "Channel name")
	debug := flag.Bool("debug", false, "Print debug?")
	flag.Parse()

	logger := log15.New()
	if *debug {
		logger.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StdoutHandler))
	} else {
		logger.SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StdoutHandler))
	}

	if *token == "" {
		logger.Crit("Token is mandatory")
		os.Exit(2)
	}

	api := slack.New(*token)
	api.SetDebug(*debug)
	channel, err := api.JoinChannel(*channelName)
	if err != nil {
		logger.Crit("Fail to join to channel")
		logger.Crit(err.Error())
		os.Exit(2)
	}

	appContext := AppContext{api, channel, *channelName, logger}

	crawler := newCrawler(time.Minute, &appContext)
	bot := newBot(crawler, &appContext)
	bot.Start()

	//forever
	<-make(chan bool)
}
