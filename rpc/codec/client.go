// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codec

import (
	"bufio"
	"fsfc/rpc/compressor"
	"fsfc/rpc/header"
	"fsfc/rpc/serializer"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"
)

// ClientCodec 为 RPC 会话的客户端实现 RPC 请求的写入和 RPC 响应的读取。
// 实现ClientCodec 接口
type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.CompressType // rpc compress type(raw,gzip,snappy,zlib) rpc压缩类型
	serializer serializer.Serializer
	response   header.ResponseHeader // rpc response header
	mutex      sync.Mutex            // protect pending map
	pending    map[uint64]string
}

// NewClientCodec Create a new client codec
// create一个客户端编解码器
func NewClientCodec(conn io.ReadWriteCloser,
	compressType compressor.CompressType, serializer serializer.Serializer) rpc.ClientCodec {

	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressType,
		serializer: serializer,
		pending:    make(map[uint64]string),
	}
}

// WriteRequest Write the rpc request header and body to the io stream
func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	// 判断压缩器是否存在
	if _, ok := compressor.Compressors[c.compressor]; !ok {
		return NotFoundCompressorError
	}
	// 用序列化器进行编码
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	// 压缩
	compressedReqBody, err := compressor.Compressors[c.compressor].Zip(reqBody)
	if err != nil {
		return err
	}
	// 从请求头部对象池取出请求头
	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()
	h.ID = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(compressedReqBody))
	h.CompressType = header.CompressType(c.compressor)
	h.Checksum = crc32.ChecksumIEEE(compressedReqBody)

	// 发送请求头
	if err := sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}
	// 发送请求体
	if err := write(c.w, compressedReqBody); err != nil {
		return err
	}

	c.w.(*bufio.Writer).Flush()
	return nil
}

// ReadResponseHeader read the rpc response header from the io stream
func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	// 重置clientCodec的响应头部
	c.response.ResetHeader()
	// 读取请求头字节串
	data, err := recvFrame(c.r)
	if err != nil {
		return err
	}
	// 用序列化器继续解码
	err = c.response.Unmarshal(data)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	// 填充 r.Seq
	r.Seq = c.response.ID
	// 填充 r.Error
	r.Error = c.response.Error
	// 根据序号填充 r.ServiceMethod
	r.ServiceMethod = c.pending[r.Seq]
	// 删除pending里的序号
	delete(c.pending, r.Seq)
	c.mutex.Unlock()
	return nil
}

// ReadResponseBody read the rpc response body from the io stream
func (c *clientCodec) ReadResponseBody(param interface{}) error {
	if param == nil {
		// 废弃多余部分
		if c.response.ResponseLen != 0 {
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
		}
		return nil
	}

	// 根据响应体长度，读取该长度的字节串
	respBody := make([]byte, c.response.ResponseLen)
	err := read(c.r, respBody)
	if err != nil {
		return err
	}

	// 校验
	if c.response.Checksum != 0 {
		if crc32.ChecksumIEEE(respBody) != c.response.Checksum {
			return UnexpectedChecksumError
		}
	}

	// 判断压缩器是否存在
	if _, ok := compressor.Compressors[c.response.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}

	// 解压
	resp, err := compressor.Compressors[c.response.GetCompressType()].Unzip(respBody)
	if err != nil {
		return err
	}

	// 反序列化
	return c.serializer.Unmarshal(resp, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
