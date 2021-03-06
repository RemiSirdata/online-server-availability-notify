package main

import (
	"time"
	"fmt"
)

type Server struct {
	Name                 string
	Range                string
	ServerAvailable      int
	PreviousAvailability int
	LastUpdate           time.Time
}

func (s *Server) GetAvailabilityMessage() string {
	if s.PreviousAvailability >= s.ServerAvailable {
		return fmt.Sprintf(MESSAGE_SERVER_AVAILABILITY_REDUCE, s.Name, s.ServerAvailable, s.PreviousAvailability)
	}
	return fmt.Sprintf(MESSAGE_SERVER_AVAILABILITY, s.Name, s.ServerAvailable, s.PreviousAvailability)
}

func (s *Server) GetStatus() string {
	return fmt.Sprintf(MESSAGE_SERVER_STATUS, s.Name, s.ServerAvailable)
}

type ServerList struct {
	Servers map[string]*Server
}

func newServerList() ServerList {
	return ServerList{
		map[string]*Server{},
	}
}

// Return server and if exist
func (s *ServerList) GetServerByName(name string) (*Server, bool) {
	if server, ok := s.Servers[name]; ok {
		return server, true
	}
	return &Server{}, false
}

func (s *ServerList) AddServer(name string, serverRange string, availability int) *Server {
	server := Server{
		name, serverRange, availability, availability, time.Now(),
	}
	s.Servers[name] = &server
	return &server
}
