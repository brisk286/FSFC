package main

import (
	"fmt"
	"fsfc/config"
	"fsfc/rpc/codec"
	"fsfc/rpc/compressor"
	"fsfc/rpc/data_rpc/protocol"
	"fsfc/rpc/serializer"
	"fsfc/rsync"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"strings"
)

// Client rpc client based on net/rpc implementation
type Client struct {
	*rpc.Client
}

//Option provides options for rpc
type Option func(o *options)

type options struct {
	compressType compressor.CompressType
	serializer   serializer.Serializer
}

// WithCompress set client compression format
func WithCompress(c compressor.CompressType) Option {
	return func(o *options) {
		o.compressType = c
	}
}

// WithSerializer set client serializer
func WithSerializer(serializer serializer.Serializer) Option {
	return func(o *options) {
		o.serializer = serializer
	}
}

// NewClient Create a new rpc client
func NewClient(conn io.ReadWriteCloser, opts ...Option) *Client {
	options := options{
		compressType: compressor.Raw,
		serializer:   serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}
	return &Client{
		rpc.NewClientWithCodec(
			codec.NewClientCodec(conn, options.compressType, options.serializer)),
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

func Rsync(changedFiles []string) {
	//conn, err := net.Dial("tcp", ":8008")
	conn, err := net.Dial("tcp", config.Config.Web.RemoteIp+config.Config.Web.RemotePort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn)
	defer client.Close()

	fileBlockHashes := client.Rpc1(changedFiles)

	for _, blockHashes := range fileBlockHashes {
		filename := blockHashes.Filename
		relaPath := AbsToRela(strings.ReplaceAll(filename, "/", "\\"))
		//localPath := FixDir(config.GetConfig().Set.RemotePath)
		localPath := FixDir(config.Config.Set.LocalPath)
		fmt.Println("relaPath:" + relaPath)
		fmt.Println("localPath:" + localPath)
		absPath := localPath + relaPath

		modified, err := ioutil.ReadFile(absPath)
		if err != nil {
			fmt.Println(absPath)
			panic("读取本地文件失败")
		}
		fmt.Println("成功找到本地文件:", absPath)

		var hashRsync []rsync.BlockHash
		for _, hash := range blockHashes.BlockHashes {
			hashRsync = append(hashRsync, rsync.BlockHash{
				Index:      int(hash.Index),
				StrongHash: hash.StrongHash,
				WeakHash:   hash.WeakHash,
			})
		}

		rsyncOps := rsync.CalculateDifferences(modified, hashRsync)
		//fmt.Println(rsyncOps)
		fmt.Println("对比差异完成, 发送RsyncOps")

		rsyncOpsReq := rsync.RsyncOpsReq{
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
	err := c.Call(cases.serviceMenthod, cases.arg, reply)
	if err != nil {
		log.Fatal(err)
	}

	return reply.FileBlockHashes
}

func (c *Client) Rpc2(rsyncOpsReq rsync.RsyncOpsReq) error {
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
	reply := &protocol.Rpc1Response{}
	err := c.Call(cases.serviceMenthod, cases.arg, reply)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// AbsToRela 如果找不到，可能是lastDir，传文件名
func AbsToRela(absPath string) string {
	var RelaPath string

	lastDir := "\\" + GetLastDir(config.Config.Set.LocalPath) + "\\"

	if strings.Index(absPath, lastDir) != -1 {
		RelaPath = absPath[strings.Index(absPath, lastDir)+1:]
	} else {
		seqList := strings.Split(absPath, "\\")
		RelaPath = seqList[len(seqList)-1]
	}
	return RelaPath
}

func FixDir(localPath string) string {
	lastDir := GetLastDir(localPath)
	return localPath[:len(localPath)-len(lastDir)]
}

func GetLastDir(path string) string {
	seqList := strings.Split(path, "\\")
	lastDir := seqList[len(seqList)-1]

	return lastDir
}

// PostChangedFile 扫描修改的文件夹
//func PostChangedFile() {
//opr（rpc：store端`CalculateBlockHashes`，修改文件名，hash list

//primfs := fs.GetFs()
//changedFiles := primfs.GetChangedFile()
//
//if changedFiles == nil {
//	//fmt.Println(time.Now(), "未检测到文件修改")
//	return
//}
//fmt.Println(time.Now(), "检测到修改的文件：", changedFiles)
//
//changedFilesJson, _ := json.Marshal(changedFiles)
//
//remoteIp := config.GetConfig().Web.RemoteIp
//remotePort := config.GetConfig().Web.RemotePort
//
//resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/v1/changedFile",
//	"application/json", bytes.NewBuffer(changedFilesJson))
//if err != nil {
//	fmt.Println(err)
//	fmt.Println("changedFiles发送失败，请检查网络连接，及存储端后台是否开启")
//	fmt.Println("本次结果将存储到缓存中，将在通信成功后再次发送")
//
//	return
//}

//body, _ := ioutil.ReadAll(resp.Body)
//blockHashesReps := response.BlockHashesReps{}
//err = json.Unmarshal(body, &blockHashesReps)
//if err != nil {
//	fmt.Println(err)
//	return
//}

//fileBlockHashes := blockHashesReps.Data
//for _, blockHashes := range fileBlockHashes {
//	filename := blockHashes.Filename
//	relaPath := fs.AbsToRela(strings.ReplaceAll(filename, "/", "\\"))
//	//localPath := fs.FixDir(config.GetConfig().Set.RemotePath)
//	localPath := fs.FixDir(config.GetConfig().Set.LocalPath)
//	fmt.Println("relaPath:" + relaPath)
//	fmt.Println("localPath:" + localPath)
//	absPath := localPath + relaPath
//
//	modified, err := ioutil.ReadFile(absPath)
//	if err != nil {
//		fmt.Println(absPath)
//		panic("读取本地文件失败")
//	}
//	fmt.Println("成功找到本地文件:", absPath)
//
//	rsyncOps := rsync.CalculateDifferences(modified, blockHashes.BlockHashes)
//	//fmt.Println(rsyncOps)
//	fmt.Println("对比差异完成, 发送RsyncOps")
//
//	rsyncOpsReq := request.RsyncOpsReq{
//		Filename:       filename,
//		RsyncOps:       rsyncOps,
//		ModifiedLength: int32(len(modified)),
//	}

//var rsyncOpsProto []*protocol.RSyncOp
//for _, each := range rsyncOps {
//	rsyncOpsProto = append(rsyncOpsProto, &protocol.RSyncOp{
//		OpCode:     each.OpCode,
//		Data:       each.Data,
//		BlockIndex: each.BlockIndex,
//	})
//}
//
//bytesProto, _ := proto.Marshal(&protocol.RsyncOpsReq{
//	Filename:       filename,
//	RsyncOps:       rsyncOpsProto,
//	ModifiedLength: int32(len(modified)),
//})
//
//var Receive protocol.RsyncOpsReq
//
//todo: 将byte保存到redis
//fmt.Println("bytes:", bytesProto)
//protobuf解码
//err = proto.Unmarshal(bytesProto, &Receive)
//if err != nil {
//	panic(err)
//}
//fmt.Println("Receive:", &Receive)
//fmt.Println("rsyncOpsReq:", rsyncOpsReq)
//		PostRsyncOps(rsyncOpsReq)
//	}
//
//	defer resp.Body.Close()
//
//}

//func PostRsyncOps(rsyncOpsReq request.RsyncOpsReq) {
//	rsyncOpsJson, _ := json.Marshal(rsyncOpsReq)
//
//	remoteIp := config.GetConfig().Web.RemoteIp
//	remotePort := config.GetConfig().Web.RemotePort
//
//	resp, err := http.Post("http://"+remoteIp+":"+remotePort+"/v1/rebuildFile", "application/json", bytes.NewBuffer(rsyncOpsJson))
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	successCodeMsg := response.ResponseMsg{}
//	err = json.Unmarshal(body, &successCodeMsg)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	if successCodeMsg.Msg == "SUCCESS" {
//		fmt.Println("文件同步成功")
//	}
//
//	defer resp.Body.Close()
//
//}
//}
