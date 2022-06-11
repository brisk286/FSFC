package rsync

import (
	"crypto/md5"
)

//常量
const (
	// BLOCK 整块数据
	BLOCK = iota
	// DATA 单独修改数据
	DATA
)

const (
	// BlockSize 默认块大小
	//BlockSize = 1024 * 644
	BlockSize = 2
	// M 65536 弱哈希算法取模
	M = 1 << 16
)

type FileBlockHashes struct {
	Filename    string
	BlockHashes []BlockHash
}

// BlockHash hash块结构
type BlockHash struct {
	//哈希块下标
	Index int
	//强哈希值
	StrongHash []byte
	//弱哈希值
	WeakHash uint32
}

// RSyncOp An rsync operation (typically to be sent across the network). It can be either a block of raw data or a block index.
//rsync数据体
type RSyncOp struct {
	//操作类型
	OpCode int32 `json:"opCode"`
	//如果是DATA 那么保存数据
	Data []byte `json:"data"`
	//如果是BLOCK 保存块下标
	BlockIndex int32 `json:"blockIndex"`
}

// Returns the smaller of a or b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Returns a weak hash for a given block of data.
//弱hash
func weakHash(v []byte) (uint32, uint32, uint32) {
	var a, b uint32
	for i := range v {
		a += uint32(v[i])
		b += (uint32(len(v)-1) - uint32(i) + 1) * uint32(v[i])
	}
	return (a % M) + (1 << 16 * (b % M)), a % M, b % M
}

// Searches for a given strong hash among all strong hashes in this bucket.
//从hash块队列中遍历每个块的强hash值  一一比对
func searchStrongHash(l []BlockHash, hashValue []byte) (bool, *BlockHash) {
	for _, blockHash := range l {
		if string(blockHash.StrongHash) == string(hashValue) {
			return true, &blockHash
		}
	}
	return false, nil
}

// Returns a strong hash for a given block of data
func strongHash(v []byte) []byte {
	h := md5.New()
	h.Write(v)
	return h.Sum(nil)
}

func CalculateDifferences(content []byte, hashes []BlockHash) []RSyncOp {

	var rsyncOps []RSyncOp

	//构建一个哈希map，<下标，哈希块列表>？ 链表结构？
	hashesMap := make(map[uint32][]BlockHash)
	//defer close(opsChannel)

	//遍历每个哈希块数组
	for _, h := range hashes {
		key := h.WeakHash
		//用弱hash做key，值为哈希块
		//数组+链表！！todo：Test
		hashesMap[key] = append(hashesMap[key], h)
	}

	//移动下标  前一个匹配块的尾部
	var offset, previousMatch int
	//弱hash 3个数值
	var aweak, bweak, weak uint32
	//标记
	var dirty, isRolling bool

	for offset < len(content) {
		//一个块的尾部
		endingByte := min(offset+BlockSize, len(content)-1)
		block := content[offset:endingByte]
		//如果不用rolling
		if !isRolling {
			//弱hash的三个值
			weak, aweak, bweak = weakHash(block)
			//如果没找到对应的块  下一次进行rolling
			isRolling = true
			//如果一直找不到会一直rolling，直到找个能对应的块，两个能对应的块之间都是DATA
		} else {
			//rolling操作 计算下一个step 1 的hash值
			aweak = (aweak - uint32(content[offset-1]) + uint32(content[endingByte-1])) % M
			bweak = (bweak - (uint32(endingByte-offset) * uint32(content[offset-1])) + aweak) % M
			weak = aweak + (1 << 16 * bweak)
		}
		//如果在hashmap中找到了弱hash对应的块， 弱hash找用hashmap
		if l := hashesMap[weak]; l != nil {
			//强hash找用遍历
			blockFound, blockHash := searchStrongHash(l, strongHash(block))
			//如果从hash块队列中找到了强hash块
			if blockFound {
				//如果是DATA
				if dirty {
					//将一个数组操作体放入操作管道中
					rsyncOps = append(rsyncOps, RSyncOp{OpCode: DATA, Data: content[previousMatch:offset]})
					dirty = false
				}
				//将一个数组操作体放入操作管道中
				rsyncOps = append(rsyncOps, RSyncOp{OpCode: BLOCK, BlockIndex: int32(blockHash.Index)})
				previousMatch = endingByte
				// 找到了就不用rolling
				isRolling = false
				offset += BlockSize
				continue
			}
		}
		//如果找不到弱hash对应的块 将下一轮搜索的块标记为DATA
		dirty = true
		//rolling
		offset++
	}

	//如果最后一个块不对应,那么把所有DATA放入
	if dirty {
		rsyncOps = append(rsyncOps, RSyncOp{OpCode: DATA, Data: content[previousMatch:]})
	}

	return rsyncOps
}
