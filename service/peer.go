package service

import (
	"bufio"
	"net"
)

type peer struct {
	connect  *net.TCPConn
	reader   *bufio.Reader
	writer   *bufio.Writer
	username string
	password string
	email    string
	ip       *net.IPAddr
	to       []string
	data     []byte
}

func (i *peer) Initialize() *peer {
	i.reader, i.writer = bufio.NewReader(i.connect), bufio.NewWriter(i.connect)
	return i
}

func (i *peer) Connected() {
	_, _ = i.writer.Write([]byte(Status220))
}

func (i *peer) Close() {
	_ = i.connect.Close()
}

func (i *peer) ReceiveByte(length int) ([]byte, error) {
	data := make([]byte, length)
	num, err := i.reader.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:num], nil
}

func (i *peer) SendMessage(data []byte) {
	_, _ = i.writer.Write(data)
	_ = i.writer.Flush()
}
