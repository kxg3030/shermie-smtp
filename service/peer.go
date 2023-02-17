package service

import (
	"bufio"
	"crypto/tls"
	"net"
)

type peer struct {
	connect   net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	scanner   *bufio.Scanner
	helloName string
	username  string
	password  string
	state     *tls.ConnectionState
}

func (i *peer) Initialize() *peer {
	i.reader, i.writer = bufio.NewReader(i.connect), bufio.NewWriter(i.connect)
	i.scanner = bufio.NewScanner(i.reader)
	return i
}

func (i *peer) connected() {
	_, _ = i.writer.Write([]byte(Status220))
}

func (i *peer) close() {
	_ = i.connect.Close()
}

func (i *peer) readline() ([]byte, error) {
	return i.reader.ReadSlice('\n')
}

func (i *peer) send(data string) {
	_, _ = i.writer.Write([]byte(data))
	_ = i.writer.Flush()
}
