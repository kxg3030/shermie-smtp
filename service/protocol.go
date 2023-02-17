package service

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

const Status354 = "354 Start mail input: end with <CRLF>.<CRLF>\r\n"
const Error421 = "421 Service unavailable, closing transmission channel\r\n"
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

const Status235 = "235 Authentication successful\r\n"
const Status220 = "220 Mysmtp Service Ready\r\n"
const Status221 = "221 Bye\r\n"
const Status250 = "250 OK\r\n"
const Status3341 = "334 VXNlcm5hbWU6\r\n"
const Status3342 = "334 UGFzc3dvcmQ6\r\n"
const Status2501 = "250 AUTH LOGIN PLAINOK\r\n250-AUTH=LOGIN PLAIN\r\n250-STARTTLS\r\n250 8BITMIME\r\n"

type Protocol struct {
	client   *peer
	envelope envelope
}

func (i *Protocol) unpack(command Command) {
	switch command.action {
	case "PROXY":
		i.handlePROXY(command)
		return

	case "HELO":
		i.handleHELO(command)
		return

	case "MAIL":
		i.handleMAIL(command)
		return

	case "RCPT":
		i.handleRCPT(command)
		return

	case "STARTTLS":
		i.handleSTARTTLS(command)
		return

	case "DATA":
		i.handleDATA(command)
		return

	case "RSET":
		i.handleRSET(command)
		return

	case "NOOP":
		i.handleNOOP(command)
		return

	case "QUIT":
		i.handleQUIT(command)
		return

	case "AUTH":
		i.handleAUTH(command)
		return

	case "XCLIENT":
		i.handleXCLIENT(command)
		return

	}
}

func (i *Protocol) handlePROXY(command Command) {

}

func (i *Protocol) handleHELO(command Command) {
	if len(command.fields) < 2 {
		i.client.send(Error500)
		return
	}
	i.client.helloName = command.fields[1]
	i.client.send(Status2501)
}

func (i *Protocol) handleEHLO(command Command) {

}

func (i *Protocol) handleXCLIENT(command Command) {

}

func (i *Protocol) handleAUTH(command Command) {
	if len(command.fields) < 2 {
		i.client.send(Error501)
		return
	}
	if i.client.helloName == "" {
		i.client.send(Error502)
		return
	}
	mechanism := strings.ToUpper(command.fields[1])
	i.client.send(Status3341)
	switch mechanism {
	case "LOGIN":
		if !i.client.scanner.Scan() {
			return
		}
		byteUsername, err := base64.StdEncoding.DecodeString(i.client.scanner.Text())
		if err != nil {
			i.client.send(Error502)
			return
		}
		i.client.username = string(byteUsername)
		i.client.send(Status3342)
		if !i.client.scanner.Scan() {
			return
		}
		bytePassword, err := base64.StdEncoding.DecodeString(i.client.scanner.Text())
		if err != nil {
			i.client.send(Error502)
			return
		}
		i.client.password = string(bytePassword)
		break
	default:
		i.client.send(Error502)
		break
	}
	i.client.send(Status235)
}

func (i *Protocol) handleQUIT(command Command) {
	fmt.Print(i.envelope.data)
	i.client.send(Status221)

}

func (i *Protocol) handleNOOP(command Command) {

}

func (i *Protocol) handleRSET(command Command) {

}

func (i *Protocol) handleSTARTTLS(command Command) {
	if i.client.state.HandshakeComplete {
		i.client.send(Error503)
		return
	}
	certificate, err := (&Server{}).certificate()
	if err != nil {
		i.client.send(Error503)
		return
	}
	session := tls.Server(i.client.connect, &tls.Config{
		Certificates: []tls.Certificate{
			*certificate,
		},
	})
	err = session.Handshake()
	if err != nil {
		i.client.send(Error550)
		return
	}
	state := session.ConnectionState()
	i.client.connect = session
	i.client.reader = bufio.NewReader(i.client.connect)
	i.client.writer = bufio.NewWriter(i.client.connect)
	i.client.scanner = bufio.NewScanner(i.client.connect)
	i.client.state = &state
}

func (i *Protocol) handleRCPT(command Command) {
	if i.envelope.recipients == nil {
		i.envelope.recipients = make([]string, 0)
	}
	if len(command.params) != 2 || strings.ToUpper(command.params[0]) != "TO" {
		i.client.send(Error502)
		return
	}
	if i.envelope.sender == "" {
		i.client.send(Error502)
		return
	}
	address, err := mail.ParseAddress(command.params[1])
	if err != nil {
		i.client.send(Error502)
		return
	}
	i.envelope.recipients = append(i.envelope.recipients, address.String())
	i.client.send(Status250)
}

func (i *Protocol) handleDATA(command Command) {
	if i.client.helloName == "" {
		i.client.send(Error502)
		return
	}
	if i.envelope.recipients == nil || len(i.envelope.recipients) == 0 {
		i.client.send(Error502)
		return
	}
	i.client.send(Status354)
	_ = i.client.connect.SetDeadline(time.Now().Add(time.Duration(10) * time.Second))
	data := new(bytes.Buffer)
	reader := textproto.NewReader(i.client.reader).DotReader()
	_, err := io.CopyN(data, reader, 1024*100)
	if err == io.EOF {
		i.envelope.data = data.Bytes()
		i.client.send(Status250)
		return
	}
	if err != nil {
		return
	}
	// 清空reader里面多余的数据(io.Discard是一个垃圾桶)
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return
	}
	// 超过最大值
	i.envelope.data = nil
	i.client.send(Error501)
}

func (i *Protocol) handleMAIL(command Command) {
	if len(command.params) != 2 || strings.ToUpper(command.params[0]) != "FROM" {
		i.client.send(Error500)
		return
	}
	if i.client.helloName == "" {
		i.client.send(Error502)
		return
	}
	if i.client.username == "" {
		i.client.send(Error502)
		return
	}
	if command.params[1] != "<>" {
		address, err := mail.ParseAddress(command.params[1])
		if err != nil {
			i.client.send(Error502)
			return
		}
		i.envelope.sender = address.String()
	}

	i.client.send(Status250)
}
