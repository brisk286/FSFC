package redis

import (
	"context"
	"github.com/go-redis/redis"
	"log"
)

var (
	ctx = context.Background()
	rdb *redis.Client
	err error
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
