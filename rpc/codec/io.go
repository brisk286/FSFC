// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codec

import (
	"encoding/binary"
	"io"
	"net"
)

//向io流中写入data
func sendFrame(w io.Writer, data []byte) (err error) {
	//将data byte 转换成二进制数据减少占用空间
	//定义size数组
	var size [binary.MaxVarintLen64]byte

	//如果data为空，向size写入0
	if len(data) == 0 {
		n := binary.PutUvarint(size[:], uint64(0))
		if err = write(w, size[:n]); err != nil {
			return
		}
		return
	}

	//向 w 中写入 转换好的二进制数据
	n := binary.PutUvarint(size[:], uint64(len(data)))
	if err = write(w, size[:n]); err != nil {
		return
	}
	if err = write(w, data); err != nil {
		return
	}
	return
}

//从io流中读取datasize 从而读取data
func recvFrame(r io.Reader) (data []byte, err error) {
	size, err := binary.ReadUvarint(r.(io.ByteReader))
	if err != nil {
		return nil, err
	}
	if size != 0 {
		//先确定长度然后去读
		data = make([]byte, size)
		if err = read(r, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

//写:多次将指定长度data byte数组中的数据写到 缓冲io流中
func write(w io.Writer, data []byte) error {
	for index := 0; index < len(data); {
		n, err := w.Write(data[index:])
		if _, ok := err.(net.Error); !ok {
			return err
		}
		index += n
	}
	return nil
}

//读：多次将缓存数据读到指定长度data byte数组中
func read(r io.Reader, data []byte) error {
	for index := 0; index < len(data); {
		n, err := r.Read(data[index:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		index += n
	}
	return nil
}
