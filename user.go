package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string // IP地址
	C    chan string
	conn net.Conn

	server *Server
}

// 新建用户,同时监听user channel消息的goroutine
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server, //标明当前用户属于哪个服务器
	}

	go user.ListenMessage()

	return user
}

// 用户上线业务
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnLineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已上线")
	fmt.Println("有新用户上线了！,用户地址：", this.Addr)
}

// 用户下线业务
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnLineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已下线")
	fmt.Println("当前用户已下线！,用户地址：", this.Addr)
}

// 给当前user对应的客户端发消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg[:len(msg)-1] == "who" {
		fmt.Println("收到请求:who") // 添加调试输出
		//查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnLineMap {
			// onlineMsg := fmt.Sprintf("%s 在线\n", user.Name)
			onlineMsg := user.Name + "在线\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		fmt.Println("收到消息", msg, "消息长度为", len(msg)) // 添加调试输出
		this.server.BroadCast(this, msg)
	}

}

// 监听user channel,有消息就发送给对应客户端
func (this *User) ListenMessage() {
	for {
		//拿出channel中的消息
		msg := <-this.C
		//发送给客户端
		this.conn.Write([]byte(msg + "\n"))
	}
}
