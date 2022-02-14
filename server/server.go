package server

import (
	"fsfc/config"
	"fsfc/fs"
	"fsfc/logger"
	"sync"
	"time"
)

var MyServer = NewServer()

type Server struct {
	mutex *sync.Mutex
	//Refresh chan *Client
	fs fs.Filesystem
}

func NewServer() *Server {
	return &Server{
		mutex: &sync.Mutex{},
		//Refresh: make(chan *Client),
	}
}

func (s *Server) Start() {
	logger.Logger.Info("start server", logger.Any("start server", "start server..."))
	//循环检测通道
	logger.Logger.Info("start timeTicker", logger.Any("scanGap", config.GetConfig().Set.ScanGap))

	scanGap := config.GetConfig().Set.ScanGap
	timeTickerChan := time.Tick(time.Second * time.Duration(scanGap))

	for {
		select {
		case <-timeTickerChan:
			//todo: 发送请求

			//s.Register通道中有信息, 从s.Register取出Client
			//case conn := <-s.Register:
			//
			////将s.Ungister通道的信息赋值给conn
			//case conn := <-s.Ungister:
			//
			//case message := <-s.Broadcast:
		}
	}
}
