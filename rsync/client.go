package rsync

import (
	"fmt"
	"fsfc/fs"
	"fsfc/rpc/codec"
	"fsfc/rpc/compressor"
	"fsfc/rpc/data_rpc/protocol"
	"fsfc/rpc/serializer"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"time"
)

// Client rpc client based on net/rpc implementation
type Client struct {
	*rpc.Client
}

// Option provides options for rpc
type Option func(o *Options)

type Options struct {
	compressType compressor.CompressType
	Serializer   serializer.Serializer
}

// WithCompress set client compression format
func WithCompress(c compressor.CompressType) Option {
	return func(o *Options) {
		o.compressType = c
	}
}

// WithSerializer set client serializer
func WithSerializer(serializer serializer.Serializer) Option {
	return func(o *Options) {
		o.Serializer = serializer
	}
}

// NewClient Create a new rpc client
func NewClient(conn io.ReadWriteCloser, opts ...Option) *Client {
	options := Options{
		compressType: compressor.Raw,
		Serializer:   serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}
	return &Client{
		rpc.NewClientWithCodec(
			codec.NewClientCodec(conn, options.compressType, options.Serializer)),
	}
}

// Call synchronously calls the rpc function
// 同步call
func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return c.Client.Call(serviceMethod, args, reply)
}

// AsyncCall asynchronously calls the rpc function and returns a channel of *rpc.Call
func (c *Client) AsyncCall(serviceMethod string, args interface{}, reply interface{}) chan *rpc.Call {
	return c.Go(serviceMethod, args, reply, nil).Done
}

func Rsync(changedFileInfos []fs.FilePrimInfo) {
	//conn, err := net.Dial("tcp", ":8008")
	//conn, err := net.Dial("tcp", config.Config.Web.RemoteIp+":"+config.Config.Web.RemotePort)
	conn, err := net.Dial("tcp", "124.70.57.7:8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewClient(conn)
	defer client.Close()

	var changedFiles []string
	for _, fileInfo := range changedFileInfos {
		changedFiles = append(changedFiles, fileInfo.RelaPath)
	}

	fileBlockHashes := client.Rpc1(changedFiles)

	for _, blockHashes := range fileBlockHashes {
		filename := blockHashes.Filename

		//relaPath := fs.AbsToRela(strings.ReplaceAll(filename, "/", "\\"))
		////localPath := FixDir(config.GetConfig().Set.RemotePath)
		//localPath := fs.FixDir(config.Config.Set.LocalPath)
		//fmt.Println("relaPath:" + relaPath)
		//fmt.Println("localPath:" + localPath)
		//absPath := localPath + relaPath

		var absPath string

		for _, fileInfo := range changedFileInfos {
			if filename == fileInfo.RelaPath {
				absPath = fileInfo.AbsPath
				break
			}
		}

		modified, err := ioutil.ReadFile(absPath)
		if err != nil {
			fmt.Println(absPath)
			panic("读取本地文件失败")
		}
		fmt.Println("成功找到本地文件:", absPath)

		var hashRsync []BlockHash
		for _, hash := range blockHashes.BlockHashes {
			hashRsync = append(hashRsync, BlockHash{
				Index:      int(hash.Index),
				StrongHash: hash.StrongHash,
				WeakHash:   hash.WeakHash,
			})
		}

		rsyncOps := CalculateDifferences(modified, hashRsync)
		//fmt.Println(rsyncOps)
		fmt.Println("对比差异完成, 发送RsyncOps")

		rsyncOpsReq := RsyncOpsReq{
			Filename:       filename,
			RsyncOps:       rsyncOps,
			ModifiedLength: int32(len(modified)),
		}

		err = client.Rpc2(rsyncOpsReq)
		if err != nil {
			fmt.Println("同步失败")
		}
		fmt.Println("同步成功")
	}
}

// Rpc1 store计算hash list 并发送回来
// 参数：linux下的绝对路径
func (c *Client) Rpc1(changedFiles []string) []*protocol.FileBlockHash {
	cases := struct {
		client         *Client
		serviceMenthod string                // 调用的方法
		arg            *protocol.Rpc1Request // 参数: A float64, B float64
	}{
		client:         c,
		serviceMenthod: "RsyncService.CalculateBlockHashes",
		arg:            &protocol.Rpc1Request{Filenames: changedFiles},
	}

	reply := &protocol.Rpc1Response{}

	fmt.Printf("调用store端方法：%v\n", cases.serviceMenthod)

	err := c.Call(cases.serviceMenthod, cases.arg, reply)
	if err != nil {
		panic(err)
	}

	return reply.FileBlockHashes
}

func (c *Client) Rpc2(rsyncOpsReq RsyncOpsReq) error {
	var rsyncOpPbs []*protocol.RSyncOpPb
	for _, op := range rsyncOpsReq.RsyncOps {
		rsyncOpPbs = append(rsyncOpPbs, &protocol.RSyncOpPb{
			OpCode:     op.OpCode,
			Data:       op.Data,
			BlockIndex: op.BlockIndex,
		})
	}

	cases := struct {
		client         *Client
		serviceMenthod string                // 调用的方法
		arg            *protocol.Rpc2Request // 参数: A float64, B float64
	}{
		client:         c,
		serviceMenthod: "RsyncService.CalculateRSyncOps",
		arg: &protocol.Rpc2Request{
			Filename:       rsyncOpsReq.Filename,
			RsyncOpPbs:     rsyncOpPbs,
			ModifiedLength: rsyncOpsReq.ModifiedLength,
		},
	}
	reply := &protocol.Rpc2Response{}
	fmt.Printf("调用store端方法：%v\n", cases.serviceMenthod)
	err := c.Call(cases.serviceMenthod, cases.arg, reply)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func FileSync(fileInfos []fs.FilePrimInfo) {
	if len(fileInfos) == 0 {
		fmt.Println("无文件修改")
	} else {
		fmt.Println(time.Now(), "检测文件修改")
		Rsync(fileInfos)
	}
}
