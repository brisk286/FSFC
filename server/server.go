package server

import (
	"fsfc/client"
	"fsfc/config"
	"fsfc/fs"
	"sync"
	"time"
)

var MyServer = NewServer()

type Server struct {
	mutex *sync.Mutex
	fs    fs.Filesystem
}

func NewServer() *Server {
	return &Server{
		mutex: &sync.Mutex{},
		//Refresh: make(chan *Client),
	}
}

func (s *Server) Start() {
	//定时任务
	scanGap := config.GetConfig().Set.ScanGap
	timeTickerChan := time.Tick(time.Second * time.Duration(scanGap))

	for {
		select {
		case <-timeTickerChan:
			client.PostChangedFile()
		}
	}
}
