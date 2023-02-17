package service

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/kxg3030/shermie-smtp/utils"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	clients  []*peer
	port     int
	listener net.Listener
	state    tls.ConnectionState
	protocol *Protocol
	group    sync.WaitGroup
	backlog  chan struct{}
}

type Command struct {
	name   string
	fields []string
	line   string
	action string
	params []string
}

func NewServer(port int) *Server {
	return &Server{
		port:     port,
		protocol: &Protocol{},
		// 同时允许最大处理100个客户端的请求
		backlog: make(chan struct{}, 1),
		group:   sync.WaitGroup{},
	}
}

func (*Server) certificate() (*tls.Certificate, error) {
	certFile, keyFile := "./cert.crt", "./cert.key"
	certFileByte, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件错误：%w", err)
	}
	keyFileByte, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件错误：%w", err)
	}
	certBlock, _ := pem.Decode(certFileByte)
	keyBlock, _ := pem.Decode(keyFileByte)
	certificate, err := tls.X509KeyPair(pem.EncodeToMemory(certBlock), pem.EncodeToMemory(keyBlock))
	if err != nil {
		return nil, fmt.Errorf("解析证书文件错误：%w", err)
	}
	return &certificate, nil
}

func (i *Server) StartTLS() {
	address, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i.port))
	if err != nil {
		fmt.Println("解析tcp地址错误：", err.Error())
		return
	}
	certificate, err := i.certificate()
	if err != nil {
		fmt.Println("获取证书文件失败：", errors.Unwrap(err).Error())
		return
	}
	listener, err := tls.Listen("tcp", address.String(), &tls.Config{
		Certificates: []tls.Certificate{*certificate},
	})
	if err != nil {
		fmt.Println("启动服务失败：", err.Error())
		return
	}
	i.listener = listener
	for n := 0; n < 5; n++ {
		go i.listen()
	}
	fmt.Println("服务已启动")
	select {}
}

func (i *Server) listen() {
	for {
		conn, err := i.listener.Accept()
		if err != nil {
			fmt.Printf("%v", err)
			continue
		}
		i.group.Add(1)
		go func() {
			client := (&peer{connect: conn}).Initialize()
			defer client.close()
			defer i.group.Done()
			select {
			case i.backlog <- struct{}{}:
				session, ok := conn.(*tls.Conn)
				if ok {
					state := session.ConnectionState()
					client.state = &state
				}
				i.clients = append(i.clients, client)
				i.protocol.client = client
				i.protocol.envelope = &envelope{}
				i.handle(client)
				<-i.backlog
			default:
				client.send(Error500)
				return
			}
			go i.save()
		}()
	}
}

func (i *Server) save() {
	if i.protocol.envelope.data == nil {
		return
	}
	message, err := utils.Parse(bytes.NewReader(i.protocol.envelope.data))
	if err != nil {
		fmt.Println("解析消息体错误：", err.Error())
		return
	}
	attachment, _ := io.ReadAll(message.Attachments[0].Data)
	file, err := os.OpenFile("./x.xlsx", os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println("解析副本文件错误：", err.Error())
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, _ = file.Write(attachment)
}

func (i *Server) handle(client *peer) {
	client.send(Status220)
	defer client.close()
	for {
		for client.scanner.Scan() {
			line := client.scanner.Text()
			command := i.parse([]byte(line))
			i.protocol.unpack(command)
		}
		err := client.scanner.Err()
		if err == bufio.ErrTooLong {
			client.send(Error501)
			continue
		}
		break
	}
}

func (i *Server) parse(line []byte) Command {
	command := Command{}
	command.line = string(bytes.Trim(line, "\r\n"))
	command.fields = strings.Fields(command.line)
	if len(command.fields) > 0 {
		command.action = strings.ToUpper(command.fields[0])
		if len(command.fields) > 1 {
			if command.fields[1][len(command.fields[1])-1] == ':' && len(command.fields) > 2 {
				command.fields[1] = command.fields[1] + command.fields[2]
				command.fields = command.fields[0:2]
			}
			command.params = strings.Split(command.fields[1], ":")
		}
	}
	return command
}
