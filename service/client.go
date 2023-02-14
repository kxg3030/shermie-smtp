package service

import (
	"errors"
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
	server := (&peer{connect: conn.(*net.TCPConn)}).Initialize()
	// 接受指令
	data, err := server.ReceiveByte(64)
	if err != nil {
		fmt.Printf("error-3:%v", err)
		return
	}
	if string(data[:3]) != "220" {
		fmt.Printf("error-3:%v", errors.New("220状态码验证错误"))
		return
	}
	// 发送EHLO
	server.SendMessage([]byte(Ehlo))
	data, err = server.ReceiveByte(64)
	if err != nil {
		fmt.Printf("error-4:%v", err)
		return
	}
	if string(data[:3]) != "250" {
		fmt.Printf("error-5:%v", errors.New("250状态码验证错误"))
		return
	}
	// 发送AuthLogin
	server.SendMessage([]byte(Auth))
	data, err = server.ReceiveByte(64)
	if err != nil {
		fmt.Printf("error-5:%v", err)
		return
	}
	if string(data[:3]) != "334" {
		fmt.Printf("error-6:%v", errors.New("334状态码验证错误"))
		return
	}
	// 发送用户名
	server.SendMessage([]byte("kxg3030\r\n"))
	data, err = server.ReceiveByte(64)
	if err != nil {
		fmt.Printf("error-5:%v", err)
		return
	}
	if string(data[:3]) != "334" {
		fmt.Printf("error-6:%v", errors.New("334状态码验证错误"))
		return
	}
	// 发送密码
	server.SendMessage([]byte("a8268168\r\n"))

	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-7:%v", err)
		return
	}
	if string(data[:3]) != "235" {
		fmt.Printf("error-8:%v", errors.New("334状态码验证错误"))
		return
	}
	// 发送密码
	server.SendMessage([]byte("a8268168\r\n"))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-9:%v", err)
		return
	}
	if string(data[:3]) != "235" {
		fmt.Printf("error-10:%v", errors.New("235状态码验证错误"))
		return
	}
	// 发送FROM
	server.SendMessage([]byte("MAIL FROM:<2635996034@qq.com>\r\n"))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-11:%v", err)
		return
	}
	if string(data[:3]) != "235" {
		fmt.Printf("error-12:%v", errors.New("235状态码验证错误"))
		return
	}
	// 发送TO
	server.SendMessage([]byte("RCPT TO:<657920579@qq.com>\r\n"))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-13:%v", err)
		return
	}
	if string(data[:3]) != "250" {
		fmt.Printf("error-14:%v", errors.New("250状态码验证错误"))
		return
	}
	// 发送DATA
	server.SendMessage([]byte(Data))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-15:%v", err)
		return
	}
	if string(data[:3]) != "354" {
		fmt.Printf("error-16:%v", errors.New("354状态码验证错误"))
		return
	}
	// 发送DATA
	server.SendMessage([]byte("这是data\r\n"))
	// 发送IMF
	server.SendMessage([]byte("这是IMF\r\n"))
	server.SendMessage([]byte(End))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-17:%v", err)
		return
	}
	if string(data[:3]) != "250" {
		fmt.Printf("error-18:%v", errors.New("250状态码验证错误"))
		return
	}
	// 发送QUIT
	server.SendMessage([]byte(Quit))
	data, err = server.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-19:%v", err)
		return
	}
	if string(data[:3]) != "221" {
		fmt.Printf("error-221:%v", errors.New("221状态码验证错误"))
		return
	}
}
