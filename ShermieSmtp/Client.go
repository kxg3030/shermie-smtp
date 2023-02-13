package ShermieSmtp

import (
	"bufio"
	"net"
)

type Client struct {
	connect  *net.TCPConn
	reader   *bufio.Reader
	writer   *bufio.Writer
	username string
	password string
	email    string
	ip       *net.IPAddr
}

func (i *Client) Initialize() *Client {
	i.reader, i.writer = bufio.NewReader(i.connect), bufio.NewWriter(i.connect)
	return i
}

func (i *Client) Connected() {
	_, _ = i.writer.Write([]byte(Status220))
}

func (i *Client) Close() {
	_ = i.connect.Close()
}

func (i *Client) ReceiveByte(length int) ([]byte, error) {
	data := make([]byte, length)
	num, err := i.reader.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:num], nil
}

func (i *Client) SendByte(data []byte) {
	_, _ = i.writer.Write(data)
	_ = i.writer.Flush()
}