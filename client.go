package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client

}

func main() {
	client := NewClient("127.0.0.1", 8080)
	if client == nil {
		fmt.Println(">>>>>连接服务器失败.....")
		return
	}

	fmt.Println(">>>>>>连接服务器成功.....")

	// 优雅的阻塞主线程
	// 创建一个带缓冲的通道
	done := make(chan struct{}, 1)

	// 启动一个 goroutine，完成一些工作后关闭通道
	go func() {
		// 这里可以做一些处理
		// 完成后关闭通道
		done <- struct{}{}
	}()

	// 等待完成信号
	<-done // 只有当通道接收到信号时，主线程才会继续
}
