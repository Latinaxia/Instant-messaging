package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// 构造方法？Javar如是认为 通常直接使用NewServer函数就可以,不用加上(this *Server)
func NewServer(Ip string, port int) *Server {
	server := &Server{
		Ip:   Ip,
		Port: port,
	}
	return *&server
}

func (this *Server) Handler(conn net.Conn) {
	//成功接收连接时打印
	fmt.Println("connection success !")
}

// 启动TCP服务
func (this *Server) Start() {

	// 在指定的IP和端口上创建一个监听器
	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	// 记得关闭
	defer Listener.Close()

	// 接收
	for {
		// 阻塞，直到有新的连接到来
		conn, err := Listener.Accept()
		if err != nil {
			fmt.Println("Listener accecp err: ", err)
			continue
		}

		// 收到连接后，并发执行Handler
		go this.Handler(conn)
	}
}
