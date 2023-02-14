package service

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const Status354 = "354 Start mail input: end with <CRLF>.<CRLF>\r\n"
const Error421 = "421 Service unavailable, closing tranmission channel\r\n"
const Error450 = "450 Requested mail action not taken: mail box unavailable\r\n"
const Error550 = "550 Requested mail action not taken: mail box unavailable\r\n"
const Error503 = "503 Bad sequence of commands\r\n"
const Error221 = "221 Service closing transmission channel\r\n"
const Error251 = "251 User not local; will forward to\r\n"
const Error451 = "451 Requested action aborted: local error in processing\r\n"
const Error452 = "452 Requested action not taken: insufficient system storage\r\n"
const Error500 = "500 Syntax error, command unrecognized\r\n"
const Error501 = "501 Syntax error in parameters or arguments\r\n"
const Error502 = "502 Command not implemented\r\n"
const Error504 = "504 Command parameter not implemented\r\n"
const Error551 = "551 User not local; please try\r\n"
const Error552 = "552 Requested mail action aborted: exceeded storage allocation\r\n"
const Error553 = "553 Requested action not taken: mailbox name not allowed\r\n"
const Error554 = "554 Transaction failed\r\n"

const Error5031 = "503 Can't receive EHLO\r\n"
const Error5032 = "503 Can't receive AUTH LOGIN\r\n"
const Error5033 = "503 Can't receive DATA\r\n"
const Error5034 = "503 Can't receive ' . '\r\n"

const Status35 = "235 Authentication successful\r\n"
const Status220 = "220 Mysmtp Service Ready\r\n"
const Status221 = "221 Bye\r\n"
const Status250 = "250 OK\r\n"
const Status3341 = "334 VXNlcm5hbWU6\r\n"
const Status3342 = "334 UGFzc3dvcmQ6\r\n"
const Status2501 = "250 AUTH LOGIN PLAINOK\r\n250-AUTH=LOGIN PLAIN\r\n250-STARTTLS\r\n250 8BITMIME\r\n"

type server struct {
	client   []*peer
	port     int
	listener *net.TCPListener
}

func NewServer(port int) *server {
	return &server{port: port}
}

func (i *server) Start() {
	address, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i.port))
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	i.listener = listener
	for n := 0; n < 5; n++ {
		go i.Listen()
	}
	fmt.Println("smtp server start")
	select {}
}

func (i *server) Listen() {
	for {
		conn, err := i.listener.AcceptTCP()
		if err != nil {
			fmt.Printf("%v", err)
			continue
		}
		client := (&peer{connect: conn}).Initialize()
		i.client = append(i.client, client)
		go i.Handle(client)
	}
}

func (i *server) Handle(client *peer) {
	// 发送连接指令
	client.Connected()
	defer client.Close()
	// 发送成功指令
	client.SendMessage([]byte(Status220))
	data, err := client.ReceiveByte(256)
	if err != nil || string(data[:4]) != "HELO" {
		fmt.Println("valid hello")
		client.SendMessage([]byte(Error5031))
		return
	}
	data = data[:]
	// 发送验证指令
	client.SendMessage([]byte(Status2501))
	// 接受登陆请求
	data, err = client.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-2:%v", err)
		return
	}
	if string(data[:10]) != "AUTH LOGIN" {
		fmt.Println("valid auth login")
		client.SendMessage([]byte(Error5031))
		return
	}
	client.SendMessage([]byte(Status3341))
	data = data[:]
	// 读取用户名
	data, err = client.ReceiveByte(256)
	if err != nil {
		fmt.Println("read username error")
		return
	}
	client.username = string(bytes.Trim(data, "\r\n"))
	client.SendMessage([]byte(Status3342))
	data = data[:]
	// 读取密码
	data, err = client.ReceiveByte(256)
	if err != nil {
		fmt.Println("read password error")
		return
	}
	client.password = string(bytes.Trim(data, "\r\n"))
	client.SendMessage([]byte(Status35))
	data = data[:]
	// 接收源邮箱地址
	data, err = client.ReceiveByte(256)
	if err != nil {
		fmt.Printf("error-5:%v", err)
		return
	}
	if string(data[:9]) != "MAIL FROM" {
		fmt.Println("read mail from error")
		client.SendMessage([]byte(Error550))
		return
	}
	// 邮箱地址解析
	client.email = string(bytes.Trim(data, "\r\n"))
	client.ip, err = net.ResolveIPAddr("ip", strings.Split(client.email, "@")[0])
	client.SendMessage([]byte(Status250))
	data = data[:]
	// 读取目标邮箱
	for {
		data, err = client.ReceiveByte(256)
		if err != nil {
			fmt.Println("read mail to error")
			client.SendMessage([]byte(Error550))
			return
		}
		if string(data[:4]) != "RCPT" {
			fmt.Println("valid rcpt error")
			break
		}
		client.to = append(client.to, string(bytes.Trim(data, "\r\n")))
		client.SendMessage([]byte(Status250))
	}
	if string(data[:4]) != "DATA" {
		fmt.Println("valid data error")
		client.SendMessage([]byte(Error5033))
		return
	}
	// 接收邮件内容
	client.SendMessage([]byte(Status354))
	data = data[:]
	for {
		frame, err := client.ReceiveByte(1024)
		if err != nil {
			fmt.Println("read data error")
			client.SendMessage([]byte(Error550))
			return
		}
		lastIndex := bytes.Index(frame, []byte(".\r\n"))
		if lastIndex >= 0 {
			client.data = append(client.data, frame[:lastIndex]...)
			data = frame[lastIndex:]
			break
		}
		client.data = append(client.data, frame...)
	}
	if string(bytes.Trim(data, "\r\n")) == "." {
		client.SendMessage([]byte(Status250))
	}
	data = data[:]
	// 接收QUIT
	_, _ = client.ReceiveByte(256)
	client.SendMessage([]byte(Status221))
	fmt.Println(string(client.data))
}
