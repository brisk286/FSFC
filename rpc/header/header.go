// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package header

import (
	"encoding/binary"
	"errors"
	"fsfc/rpc/compressor"
	"sync"
)

const (
	// Varint 变长编码
	// MaxHeaderSize = 2 + 10 + 10 + 10 + 4 (10 refer to binary.MaxVarintLen64)
	MaxHeaderSize = 36

	Uint32Size = 4
	Uint16Size = 2
)

var UnmarshalError = errors.New("an error occurred in Unmarshal")

// CompressType type of compressions supported by rpc
type CompressType uint16

// RequestHeader request header structure looks like:
// +--------------+----------------+----------+------------+----------+
// | CompressType |      Method    |    ID    | RequestLen | Checksum |
// +--------------+----------------+----------+------------+----------+
// |    uint16    | uvarint+string |  uvarint |   uvarint  |  uint32  |
// +--------------+----------------+----------+------------+----------+
type RequestHeader struct {
	sync.RWMutex
	CompressType CompressType // 压缩类型
	Method       string       // 方法名
	ID           uint64       // 请求ID
	RequestLen   uint32       // 请求体长度
	Checksum     uint32       // 请求体校验 使用CRC32摘要算法
}

// Marshal will encode request header into a byte slice
func (r *RequestHeader) Marshal() []byte {
	//上读锁
	r.RLock()
	defer r.RUnlock()
	idx := 0
	// MaxHeaderSize = 2 + 10 + len(string) + 10 + 10 + 4
	// 加上method的长度
	header := make([]byte, MaxHeaderSize+len(r.Method))
	// 写入uint16类型的压缩类型
	// 强制转换，将r.CompressType写入header[idx:]，以小端序的形式
	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += writeString(header[idx:], r.Method)
	idx += binary.PutUvarint(header[idx:], r.ID)                 // 写入uvarint类型的请求ID号
	idx += binary.PutUvarint(header[idx:], uint64(r.RequestLen)) // 写入uvarint类型的请求体长度

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum) // 写入uvarint类型的校验码
	idx += Uint32Size
	return header[:idx]
}

// Unmarshal will decode request header into a byte slice
func (r *RequestHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return UnmarshalError
	}

	defer func() {
		if r := recover(); r != nil {
			err = UnmarshalError
		}
	}()
	idx, size := 0, 0
	//从data中读取字节序
	r.CompressType = CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size // 读取uint16类型的压缩类型

	//读string的字节序
	r.Method, size = readString(data[idx:])
	idx += size

	r.ID, size = binary.Uvarint(data[idx:]) // 读取uvarint类型的请求ID号
	idx += size

	length, size := binary.Uvarint(data[idx:]) // 读取uvarint类型的请求体长度
	r.RequestLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:]) // 读取uvarint类型的校验码

	return
}

// GetCompressType get compress type
func (r *RequestHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return compressor.CompressType(r.CompressType)
}

// ResetHeader reset request header
func (r *RequestHeader) ResetHeader() {
	r.Lock()
	defer r.Unlock()
	r.ID = 0
	r.Checksum = 0
	r.Method = ""
	r.CompressType = 0
	r.RequestLen = 0
}

// ResponseHeader request header structure looks like:
// +--------------+---------+----------------+-------------+----------+
// | CompressType |    ID   |      Error     | ResponseLen | Checksum |
// +--------------+---------+----------------+-------------+----------+
// |    uint16    | uvarint | uvarint+string |    uvarint  |  uint32  |
// +--------------+---------+----------------+-------------+----------+
type ResponseHeader struct {
	sync.RWMutex
	CompressType CompressType
	ID           uint64
	Error        string
	ResponseLen  uint32
	Checksum     uint32
}

// Marshal will encode response header into a byte slice
func (r *ResponseHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	header := make([]byte, MaxHeaderSize+len(r.Error)) // prevent panic

	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += writeString(header[idx:], r.Error)
	idx += binary.PutUvarint(header[idx:], uint64(r.ResponseLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

// Unmarshal will decode response header into a byte slice
func (r *ResponseHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return UnmarshalError
	}

	defer func() {
		if r := recover(); r != nil {
			err = UnmarshalError
		}
	}()
	idx, size := 0, 0
	r.CompressType = CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	r.Error, size = readString(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.ResponseLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

// GetCompressType get compress type
func (r *ResponseHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return compressor.CompressType(r.CompressType)
}

// ResetHeader reset response header
func (r *ResponseHeader) ResetHeader() {
	r.Lock()
	defer r.Unlock()
	r.Error = ""
	r.ID = 0
	r.CompressType = 0
	r.Checksum = 0
	r.ResponseLen = 0
}

func readString(data []byte) (string, int) {
	idx := 0
	length, size := binary.Uvarint(data) // 读取一个uvarint类型表示字符的长度
	idx += size
	str := string(data[idx : idx+int(length)])
	idx += len(str)
	return str, idx
}

//写入string
func writeString(data []byte, str string) int {
	idx := 0
	idx += binary.PutUvarint(data, uint64(len(str))) // 写入一个uvarint类型表示字符长度
	copy(data[idx:], str)
	idx += len(str)
	return idx
}
