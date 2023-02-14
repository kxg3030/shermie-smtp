package ShermieSmtp

import (
	"bufio"
	"net"
)

type Conn struct {
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

func (i *Conn) Initialize() *Conn {
	i.reader, i.writer = bufio.NewReader(i.connect), bufio.NewWriter(i.connect)
	return i
}

func (i *Conn) Connected() {
	_, _ = i.writer.Write([]byte(Status220))
}

func (i *Conn) Close() {
	_ = i.connect.Close()
}

func (i *Conn) ReceiveByte(length int) ([]byte, error) {
	data := make([]byte, length)
	num, err := i.reader.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:num], nil
}

func (i *Conn) SendMessage(data []byte) {
	_, _ = i.writer.Write(data)
	_ = i.writer.Flush()
}
