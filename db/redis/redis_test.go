package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"testing"
)

func init() {
	//连接redis
	rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", Password: "", DB: 0})
	//健康检测
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalln("redis状态错误: ", err)
	}
}

func Test_Kv(t *testing.T) {
	//新增k/v
	_ = rdb.Set(ctx, "hello", "world", 0).Err()

	//获取k/v
	result, _ := rdb.Get(ctx, "hello").Result()
	fmt.Println(result)

	//删除
	//_, _ = rdb.Del(ctx, "hello").Result()
}

func Test_List(t *testing.T) {
	//新增
	//_ = rdb.RPush(ctx, "list", "message").Err()
	//_ = rdb.RPush(ctx, "list", "message2").Err()

	//查询
	//result, _ := rdb.LLen(ctx, "list").Result()
	//fmt.Println(result)
	l := rdb.LLen(ctx, "list")
	fmt.Println(l)

	//更新
	//_ = rdb.LSet(ctx, "list", 2, "message set").Err()

	//result, _ = rdb.LLen(ctx, "list").Result()
	//fmt.Println(result)

	//遍历
	//lRange, _ := rdb.LRange(ctx, "list", 0, result).Result()
	//for _, v := range lRange {
	//	log.Println(v)
	//}
	//
	////删除
	//_, _ = rdb.LRem(ctx, "list", 3, "message2").Result()
}

func Test_li(t *testing.T) {
	// 返回从0开始到-1位置之间的数据，意思就是返回全部数据
	vals, err := rdb.LRange(ctx, "list", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(vals)
}
