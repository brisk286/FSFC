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

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request    header.RequestHeader
	serializer serializer.Serializer
	mutex      sync.Mutex // protects seq, pending
	seq        uint64
	pending    map[uint64]uint64
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		pending:    make(map[uint64]uint64),
	}
}

// ReadRequestHeader read the rpc request header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	// 重置serverCodec结构体的请求头部
	s.request.ResetHeader()
	// 读取请求头部字节串
	data, err := recvFrame(s.r)
	if err != nil {
		return err
	}
	//将字节串反序列化
	err = s.request.Unmarshal(data)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	// 序号自增
	s.seq++
	// 自增序号与请求头部的ID进行绑定
	s.pending[s.seq] = s.request.ID
	// 填充 r.ServiceMethod
	r.ServiceMethod = s.request.Method
	// 填充 r.Seq
	r.Seq = s.seq
	s.mutex.Unlock()
	return nil
}

// ReadRequestBody read the rpc request body from the io stream
func (s *serverCodec) ReadRequestBody(param interface{}) error {
	if param == nil {
		// 废弃多余部分
		if s.request.RequestLen != 0 {
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
		}
		return nil
	}

	reqBody := make([]byte, s.request.RequestLen)

	// 根据请求体的大小，读取该大小的字节串
	err := read(s.r, reqBody)
	if err != nil {
		return err
	}

	// 校验
	if s.request.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != s.request.Checksum {
			return UnexpectedChecksumError
		}
	}

	// 判断压缩器是否存在
	if _, ok := compressor.
		Compressors[s.request.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}

	// 解压
	req, err := compressor.
		Compressors[s.request.GetCompressType()].Unzip(reqBody)
	if err != nil {
		return err
	}

	// 把字节串反序列化
	return s.serializer.Unmarshal(req, param)
}

// WriteResponse Write the rpc response header and body to the io stream
func (s *serverCodec) WriteResponse(r *rpc.Response, param interface{}) error {
	s.mutex.Lock()
	id, ok := s.pending[r.Seq]
	if !ok {
		s.mutex.Unlock()
		return InvalidSequenceError
	}
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	// 如果RPC调用结果有误，把param置为nil
	if r.Error != "" {
		param = nil
	}

	// 判断压缩器是否存在
	if _, ok := compressor.
		Compressors[s.request.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}

	var respBody []byte
	var err error

	// 反序列化
	if param != nil {
		respBody, err = s.serializer.Marshal(param)
		if err != nil {
			return err
		}
	}

	// 压缩
	compressedRespBody, err := compressor.
		Compressors[s.request.GetCompressType()].Zip(respBody)
	if err != nil {
		return err
	}
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.ResetHeader()
		header.ResponsePool.Put(h)
	}()
	h.ID = id
	h.Error = r.Error
	h.ResponseLen = uint32(len(compressedRespBody))
	h.Checksum = crc32.ChecksumIEEE(compressedRespBody)
	h.CompressType = s.request.CompressType

	// 发送响应头
	if err = sendFrame(s.w, h.Marshal()); err != nil {
		return err
	}

	// 发送响应体
	if err = write(s.w, compressedRespBody); err != nil {
		return err
	}
	s.w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}
