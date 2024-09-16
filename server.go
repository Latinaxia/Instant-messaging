package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// 构造方法？Javar如是认为 通常直接使用NewServer函数就可以,不用加上(this *Server)
func NewServer(Ip string, port int) *Server {
	server := &Server{
		Ip:        Ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return *&server
}

// 将广播消息存入服务端channel
func (this *Server) BroadCast(user *User, msg string) {
	// sendMsg := fmt.Sprintf("[%s]:%s", user.Addr, msg)
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

//广播消息的方法,将消息发送给全体在线user
/*
注意，在 Go中，通过通道发送的值在被读取后会“消失”。
当一个值被发送到通道中时，只有在有接收方读取该值时，这个值才会被从通道中移除。
这就意味着一旦接收方成功接收到这个值，通道中的该值就不再存在
如果想多次读取这个值的话，可以采用广播的方式，即如下for循环的实现
*/
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		//将消息发送给全体在线user
		for _, user := range this.OnLineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("connection success !")

	//用户上线，加入到OnlineMap
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播用户上线消息
	this.BroadCast(user, "已上线")
	fmt.Println("有新用户上线了！,用户地址：", user.Addr)

	//阻塞住,保持用户的连接
	select {}
}

// 启动TCP服务
func (this *Server) Start() {

	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port)) // 在指定的IP和端口上创建一个监听器
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	defer Listener.Close() // 记得关闭

	go this.ListenMessage() // 启动消息监听Message,有msg就发送给全体user

	// 接收
	for {
		conn, err := Listener.Accept() // 阻塞，直到有新的连接到来
		if err != nil {
			fmt.Println("Listener accecp err: ", err)
			continue
		}

		go this.Handler(conn) // 收到连接后，并发执行Handler
	}
}