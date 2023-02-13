package ShermieSmtp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const Status2501 = "250 AUTH LOGIN PLAINOK\r\n250-AUTH=LOGIN PLAIN\r\n250-STARTTLS\r\n250 8BITMIME\r\n"
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
const Status35 = "235 Authentication successful\r\n"

const Error5031 = "503 Can't receive EHLO\r\n"
const Error5032 = "503 Can't receive AUTH LOGIN\r\n"
const Error5033 = "503 Can't receive DATA\r\n"
const Error5034 = "503 Can't receive ' . '\r\n"

const Status220 = "220 Mysmtp Service Ready\r\n"
const Status221 = "221 Bye\r\n"
const Status250 = "250 OK\r\n"
const Status3341 = "334 VXNlcm5hbWU6\r\n"
const Status3342 = "334 UGFzc3dvcmQ6\r\n"

const Auth = "AUTH LOGIN\r\n"
const Data = "DATA\r\n"
const Quit = "QUIT\r\n"
const End = "\r\n.\r\n"

type Server struct {
	client   []*Client
	port     int
	listener *net.TCPListener
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

func (i *Server) Start() {
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

func (i *Server) Listen() {
	for {
		conn, err := i.listener.AcceptTCP()
		if err != nil {
			fmt.Printf("%v", err)
			continue
		}
		client := (&Client{connect: conn}).Initialize()
		i.client = append(i.client, client)
		go i.Handle(client)
	}
}

func (i *Server) Handle(client *Client) {
	// 发送连接指令
	client.Connected()
	defer client.Close()
	for {
		// 接收EHLO指令
		data, err := client.ReceiveByte(4)
		if err != nil || string(data) != "EHLO" {
			fmt.Printf("error-1:%s\n", "ehlo验证失败")
			client.SendByte([]byte(Error5031))
			break
		}
		// 发送验证指令
		client.SendByte([]byte(Status2501))
		// 接受登陆请求
		data, err = client.ReceiveByte(1024)
		if err != nil {
			fmt.Printf("error-2:%v", err)
			continue
		}
		if string(data[:10]) != "AUTH LOGIN" {
			fmt.Printf("error-2:%v\n", err)
			client.SendByte([]byte(Error5031))
			continue
		}
		client.SendByte([]byte(Status3341))
		// 读取用户名
		data, err = client.ReceiveByte(1024)
		if err != nil {
			fmt.Printf("error-3:%v\n", err)
			continue
		}
		client.username = string(data)
		client.SendByte([]byte(Status3342))
		// 读取密码
		data, err = client.ReceiveByte(1024)
		if err != nil {
			fmt.Printf("error-4:%v\n", err)
			continue
		}
		client.password = string(data)
		client.SendByte([]byte(Status35))
		// 接收源邮箱地址
		data, err = client.ReceiveByte(1024)
		if err != nil {
			fmt.Printf("error-5:%v", err)
			continue
		}
		if string(data[:9]) != "MAIL FROM" {
			fmt.Printf("error-6:%s\n", "mail from 验证失败")
			client.SendByte([]byte(Error550))
			continue
		}
		// 验证邮箱正确
		// 邮箱地址解析
		client.email = string(data)
		client.ip, err = net.ResolveIPAddr("ip", strings.Split(client.email, "@")[0])
		client.SendByte([]byte(Status250))
	}

}
