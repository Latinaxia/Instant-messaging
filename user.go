package main

import "net"

type User struct {
	Name string
	Addr string // IP地址
	C    chan string
	conn net.Conn
}

// 新建用户,同时监听user channel消息的goroutine
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	go user.ListenMessage()

	return user
}

//监听user channel,有消息就发送给对应客户端
func (this *User) ListenMessage() {
	for {
		//拿出channel中的消息
		msg := <-this.C
		//发送给客户端
		this.conn.Write([]byte(msg + "\n"))
	}
}
