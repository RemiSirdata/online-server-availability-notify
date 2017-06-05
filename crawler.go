package main

import (
	"time"
	"net/http"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"strconv"
	"fmt"
)

type Crawler struct {
	timelaps   time.Duration
	AppContext *AppContext
	Listener   []*Listener
	ServerList ServerList

	addListener chan *Listener
	//if crawl has started (need at least one listener)
	started bool
}

func newCrawler(timelaps time.Duration, context *AppContext) *Crawler {
	if timelaps < time.Minute {
		context.Logger.Info("Timestamp too short, changed to 1 minute")
		timelaps = time.Minute
	}
	crawler := Crawler{
		timelaps:   timelaps,
		AppContext: context,
		started:    false,
		ServerList: newServerList(),
	}
	crawler.listenAddListener()
	return &crawler
}

// Avoid concurrent write event if low probability
func (c *Crawler) listenAddListener() {
	c.addListener = make(chan *Listener)
	go func() {
		for l := range c.addListener {
			c.Listener = append(c.Listener, l)
		}
	}()
}

func (c *Crawler) AddListener() *Listener {
	return c.AddListenerForServer("")
}

func (c *Crawler) AddListenerForServer(serverName string) *Listener {
	if !c.started {
		c.StartCrawl()
	}
	l := Listener{
		make(chan Server),
		serverName,
	}
	c.addListener <- &l
	return &l
}

func (c *Crawler) StartCrawl() {
	go func() {
		for {
			c.checkAvailability()
			time.Sleep(c.timelaps)
		}
	}()
}

func (c *Crawler) checkAvailability() {
	resp, err := http.Get(ONLINE_PAGE_SERVER_LIST)
	if err != nil {
		c.AppContext.Logger.Error(fmt.Sprintf("Fail to get %s : %s", ONLINE_PAGE_SERVER_LIST, err.Error()))
		return
	}
	defer resp.Body.Close()
	root, err := html.Parse(resp.Body)
	if err != nil {
		c.AppContext.Logger.Error(fmt.Sprintf("Fail to get body %s : %s", ONLINE_PAGE_SERVER_LIST, err.Error()))
		return
	}
	serverAnchorNodes := scrape.FindAll(root, scrape.ByClass("Item__link--foot"))
	for _, serverAnchorNode := range serverAnchorNodes {
		path := scrape.Attr(serverAnchorNode, "href")
		c.checkAvailabilityServer(fmt.Sprintf("http://%s%s", resp.Request.Host, path))
	}
}

func (c *Crawler) checkAvailabilityServer(url string) {
	resp, err := http.Get(url)
	if err != nil {
		c.AppContext.Logger.Crit(fmt.Sprintf("Fail to get %s : %s", url, err.Error()))
		return
	}
	root, err := html.Parse(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		c.AppContext.Logger.Error(fmt.Sprintf("Fail to get body %s : %s", url, err.Error()))
		return
	}
	serverNumberNodes := scrape.FindAll(root, scrape.ByClass("typenumber--total"))
	totalAvailable := 0
	for _, serverNumberNode := range serverNumberNodes {
		i, err := strconv.Atoi(scrape.Text(serverNumberNode))
		if err != nil {
			c.AppContext.Logger.Error(fmt.Sprintf("Fail to get availables server for url %s", url))
		} else {
			totalAvailable += i
		}
	}

	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.Parent != nil && scrape.Attr(n.Parent, "class") == "Dedidetails__maintitle Common__title Common__title--main beta--low text-center" {
			return scrape.Attr(n, "itemprop") == "name"
		}
		return false
	}
	serverNameNode, ok := scrape.Find(root, matcher)
	if !ok {
		c.AppContext.Logger.Error(fmt.Sprintf("Fail to get server name for url %s", url))
		return
	}
	serverName := scrape.Text(serverNameNode)
	c.updateAvailability(serverName, totalAvailable)
}

func (c *Crawler) updateAvailability(serverName string, totalAvailable int) {
	server, ok := c.ServerList.GetServerByName(serverName)
	if ok {
		if server.ServerAvailable != totalAvailable {
			prev := server.ServerAvailable
			server.ServerAvailable = totalAvailable
			server.PreviousAvailability = prev
			c.notifyForServer(*server)
		}
	} else {
		server := c.ServerList.AddServer(serverName, "", totalAvailable)
		c.notifyForServer(*server)
	}
}

func (c *Crawler) notifyForServer(server Server) {
	for _, listener := range c.Listener {
		if listener.ServerTypeFilter == "" || listener.ServerTypeFilter == server.Name {
			listener.Chan <- server
		}
	}
}

type Listener struct {
	Chan             chan Server
	ServerTypeFilter string
}
