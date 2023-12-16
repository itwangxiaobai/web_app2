package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"web_app/settings"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

// Init 初始化连接
func Init(cfg *settings.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			//viper.GetString("redis.host"),
			//viper.GetInt("redis.port"),
			cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	_, err = rdb.Ping().Result()
	return
}

func Close() {
	_ = rdb.Close()
}
