package router

import (
	"fsfc/fs"
	"fsfc/rpc/codec"
	"fsfc/rpc/serializer"
	"fsfc/rsync"
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

func NewServer(opts ...rsync.Option) *Server {
	options := rsync.Options{
		Serializer: serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}

	return &Server{&sync.Mutex{}, &rpc.Server{}, options.Serializer, fs.PrimFs}

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

// Start 同步部分
func (s *Server) Start() {
	go s.TickerScan()
	go s.TickerMlScan()
}

// TickerScan 手动同步部分：
// 循环检测文件修改情况
func (s *Server) TickerScan() {
	timeTickerChan := time.Tick(time.Second * time.Duration(s.fs.ScanGap))

	for {
		select {
		case <-timeTickerChan:
			rsync.FileSync(fs.PrimFs.GetChangedFile())
		}
	}
}

// TickerMlScan ML自动同步部分：
// 维护一个动态文件列表，设定size
//
// 每隔1个量化单位【1】，更新动态文件列表，通过扫描Recent文件夹文件修改情况
// 每隔1个量化单位【2】（时间应该远远大于手动量化时间【1】）进行一次TopK同步，同步过程：将列表中文件信息放入模型，得到预测结果：是否同步
func (s *Server) TickerMlScan() {
	timeTickerUpdateChan := time.Tick(time.Second * time.Duration(s.fs.ScanMlGapUpdate))
	timeTickerSyncChan := time.Tick(time.Second * time.Duration(s.fs.ScanMlGapSync))

	for {
		select {
		case <-timeTickerUpdateChan:
			fs.PrimFs.UpdateFileInfo()
		case <-timeTickerSyncChan:
		priority:
			for {
				select {
				case <-timeTickerSyncChan:
					fs.PrimFs.UpdateFileInfo()
				default:
					break priority
				}
			}
			//fs.PrimFs.CollectData()
			rsync.FileSync(fs.PrimFs.GetMlChangedFile())
		}
	}

}
