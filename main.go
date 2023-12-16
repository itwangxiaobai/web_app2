package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web_app/dao/mysql"
	"web_app/dao/redis"
	"web_app/logger"
	"web_app/routes"
	"web_app/settings"
)

// Go web开发较通用的脚手架模板

func main() {
	// 1. 加载配置
	// 使用os.args获取命令行参数
	//if len(os.Args) < 2 {
	//	fmt.Println("need config file.eg: bluebell config.yaml")
	//	return
	//}
	//if err := settings.Init(os.Args[1]); err != nil {
	//	fmt.Printf("init settings failed, err:%v\n", err)
	//	return
	//}

	// 使用flag获取命令行参数
	var configpath string
	flag.StringVar(&configpath, "configpath", "./config.yaml", "配置文件路径...")
	flag.Parse()
	fmt.Println(configpath)
	if err := settings.Init(configpath); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	fmt.Println(settings.Conf)
	//fmt.Println(settings.Conf.LogConfig == nil)
	// 2. 初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success...")
	// 3. 初始化Mysql连接
	if err := mysql.Init(settings.Conf.MysqlConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close()
	// 4. 初始化Redis连接
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()
	// 5. 注册路由
	r := routes.SetUp(settings.Conf.Mode)
	// 6. 启动服务（优雅关机）
	fmt.Println(settings.Conf.Port)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()
	// 等待信号中断来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认发送 syscall,SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的ctrl+c就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的信号 syscall.SIGINT 和 syscall.SIGTERM 信号发送给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("shutdown Server...")
	// 创建一个5秒超时时间的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务，（将未处理完的请求处理完再关闭服务） 超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
