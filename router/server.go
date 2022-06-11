package router

import (
	"fsfc/config"
	"fsfc/fs"
	"fsfc/rpc/codec"
	"fsfc/rpc/serializer"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

var MyServer = NewServer()

// Server rpc server based on net/rpc implementation
type Server struct {
	mutex *sync.Mutex
	*rpc.Server
	serializer.Serializer
	fs fs.Filesystem
}

func NewServer(opts ...Option) *Server {
	options := options{
		serializer: serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}

	return &Server{&sync.Mutex{}, &rpc.Server{}, options.serializer, fs.GetFs()}

}

// Register register rpc function
func (s *Server) Register(rcvr interface{}) error {
	return s.Server.Register(rcvr)
}

// RegisterName register the rpc function with the specified name
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.Server.RegisterName(name, rcvr)
}

// Serve start service
func (s *Server) Serve(lis net.Listener) {
	log.Printf("tinyrpc started on: %s", lis.Addr().String())
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn, s.Serializer))
	}
}

func (s *Server) Start() {
	//定时任务
	scanGap := config.GetConfig().Set.ScanGap
	timeTickerChan := time.Tick(time.Second * time.Duration(scanGap))

	for {
		select {
		case <-timeTickerChan:
			PostChangedFile()
		}
	}
}
