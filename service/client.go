package service

import (
	"fmt"
	"net"
	"strconv"
)

const Ehlo = "EHLO\r\n"
const Auth = "AUTH LOGIN\r\n"
const Data = "DATA\r\n"
const Quit = "QUIT\r\n"
const End = "\r\n.\r\n"

type Client struct {
	port int
}

func (i *Client) NewClient(port int) *Client {
	return &Client{
		port: port,
	}
}

func (i *Client) Connect() {
	addr, err := net.ResolveIPAddr("tcp", ":"+strconv.Itoa(i.port))
	if err != nil {
		fmt.Printf("error-1:%v", err)
		return
	}
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		fmt.Printf("error-2:%v", err)
		return
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	//server := (&peer{connect: conn.(*net.TCPConn)}).Initialize()
	// 接受指令

}
