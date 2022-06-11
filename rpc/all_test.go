package tinyrpc

import (
	"encoding/json"
	"errors"
	"fsfc/router"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"testing"

	js "fsfc/rpc/test.data/json"
	pb "fsfc/rpc/test.data/message"
	"github.com/stretchr/testify/assert"
)

func init() {
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}

	//起一个server
	server := router.NewServer()
	//server注册器
	err = server.Register(new(pb.ArithService))
	if err != nil {
		log.Fatal(err)
	}

	go server.Serve(lis)

	lis, err = net.Listen("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}

	//第二个server
	server = router.NewServer(router.WithSerializer(&Json{}))
	err = server.Register(new(js.TestService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)
}

//同步call测试
// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	//起一个client
	client := router.NewClient(conn)
	defer client.Close()

	//expect结构体，包含返回值与err
	type expect struct {
		reply *pb.ArithResponse // reply用pb协议编码 : C float64
		err   error
	}
	//初始化测试用例，匿名结构体
	cases := []struct {
		client         *router.Client
		name           string
		serviceMenthod string           // 调用的方法
		arg            *pb.ArithRequest // 参数: A float64, B float64
		expect         expect           // expect结构体：reply和err
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 15},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 100},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&pb.ArithRequest{A: 20, B: 0},
			expect{
				&pb.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			//client.Call()
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

//异步call测试
// TestClient_AsyncCall test client asynchronously call
func TestClient_AsyncCall(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn)
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 15},
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 100},
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&pb.ArithRequest{A: 20, B: 0},
			expect{
				&pb.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			call := c.client.AsyncCall(c.serviceMenthod, c.arg, reply)
			err := <-call
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err.Error)
		})
	}
}

//snappy压缩测试
// TestNewClientWithSnappyCompress test snappy comressor
func TestNewClientWithSnappyCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

//gzip压缩测试
// TestNewClientWithGzipCompress test gzip comressor
func TestNewClientWithGzipCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

//zlib压缩测试
// TestNewClientWithZlibCompress test zlib compressor
func TestNewClientWithZlibCompress(t *testing.T) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(compressor.Gzip))
	defer client.Close()

	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

//注册器测试
// TestServer_Register .
func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.RegisterName("ArithService", new(pb.ArithService))
	assert.Equal(t, nil, err)
	err = server.Register(new(pb.ArithService))
	assert.Equal(t, errors.New("rpc: service already defined: ArithService"), err)
}

// Json .
type Json struct{}

// Marshal .
func (_ *Json) Marshal(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// Unmarshal .
func (_ *Json) Unmarshal(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

//序列化器测试
// TestNewClientWithSerializer .
func TestNewClientWithSerializer(t *testing.T) {

	conn, err := net.Dial("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithSerializer(&Json{}))
	defer client.Close()

	type expect struct {
		reply *js.Response
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *js.Request
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "TestService.Add",
			arg:            &js.Request{A: 20, B: 5},
			expect: expect{
				reply: &js.Response{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &js.Response{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}
