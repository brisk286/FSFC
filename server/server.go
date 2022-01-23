package server

//
//import (
//	"go-chat/pkg/global/log"
//	"sync"
//)
//
//var MyServer = NewServer()
//
//type Server struct {
//	mutex   *sync.Mutex
//	Refresh chan *Client
//}
//
//func NewServer() *Server {
//	return &Server{
//		mutex:   &sync.Mutex{},
//		Refresh: make(chan *Client),
//	}
//}
//
//func (s *Server) Start() {
//	log.Logger.Info("start server", log.Any("start server", "start server..."))
//	//循环检测三个通道
//	for {
//		select {
//		//s.Register通道中有信息, 从s.Register取出Client
//		case conn := <-s.Register:
//
//		//将s.Ungister通道的信息赋值给conn
//		case conn := <-s.Ungister:
//
//		case message := <-s.Broadcast:
//		}
//	}
//}
