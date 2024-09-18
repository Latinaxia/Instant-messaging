package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
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

// conn传值
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("connection success !")

	//用户上线，加入到OnlineMap
	user := NewUser(conn, this)
	user.Online()

	//监听user是否活跃的channel
	isLive := make(chan bool)

	//接收客户端发送的消息
	go func() {
		var msgBuffer string // 使用一个字符串来缓冲消息
		buf := make([]byte, 4096)
		for {
			//阻塞调用，如果连接关闭后，仍然有可能会继续处理读取的逻辑，因为 Read 方法不会立即返回
			n, err := conn.Read(buf)

			if err != nil {
				if err == io.EOF {
					// 连接正常关闭，处理下线
					user.Offline()
					return // 直接返回退出goroutine
				}
				fmt.Println("Conn Read err: ", err)
				return // 遇到其他错误也退出goroutine
			}

			if n == 0 {
				// 处理空读取，通常表示连接关闭
				user.Offline()
				return
			}

			if err != nil && err != io.EOF { //io.EOF 是一个特殊的错误，表示已经到达文件或流的末尾。在网络编程中，这通常表示连接已被对方关闭，是一种正常的情况。
				fmt.Println("Conn Read err: ", err)
			}

			msgBuffer += string(buf[:n]) // 将读取的内容累加到消息缓冲区

			// 检查是否包含换行符（回车的情况）
			if len(msgBuffer) > 0 && msgBuffer[len(msgBuffer)-1] == '\n' {
				// this.BroadCast(user, msgBuffer)
				/*
				   去掉字符串两端的空白字符，包括空格、制表符（tab）、换行符（\n）、回车符（\r）等。
				   接受的msg末尾有"\r"回车符
				   通常是在Windows系统的控制台输入中按下回车键（Enter）后会产生的结果。
				   在这种情况下，msg后面是 \r 而不是 \n，所以导致字符串的输出中出现额外的字符。
				*/
				user.DoMessage(strings.TrimSpace(msgBuffer))
				msgBuffer = "" // 清空缓冲区

				//用户的任意消息，代表当前用户是活跃的
				isLive <- true
			}
		}
	}()

	//阻塞住,保持用户的连接
	for {
		select {
		case <-isLive:
			//当前用户活跃，重置定时器
		case <-time.After(time.Second * 10):
			//已经超时，将当前user强制关闭
			user.SendMsg("你被踢了")
			user.Offline()
			// close(user.C)
			// conn.Close()
			return
		}
	}

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
			fmt.Println("Listener accept err: ", err)
			continue
		}

		go this.Handler(conn) // 收到连接后，并发执行Handler
	}
}
