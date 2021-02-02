package gof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"time"
)

type ConnStatus int

const (
	CONN_NEW     ConnStatus = 1 //新连接
	CONN_CLOSE   ConnStatus = 2 //关闭连接
	CONN_MESSAGE ConnStatus = 3 //处理消息
)

// Close codes defined in RFC 6455, section 11.7.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

type Conn struct {
	Server          *Server
	fd              int       //当前连接的文件描述符 fd
	UpdateTime      time.Time //最新的更新时间，判断超时用
	ReadBufferSize  int       // 读内容的默认大小
	WriteBufferSize int       //写内容的默认大小
	ReadBuf         chan Message
	WriteBuf        chan Message //用于
	Method          string       //请求方式 websocket必须是get请求方式
	CloseCode       uint16
	CloseReason     []byte
}

type Message struct {
	Conn        *Conn
	MessageType int
	Content     []byte
}

func newConn(fd, ReadBufferSize, WriteBufferSize int, server *Server) *Conn {
	return &Conn{
		fd:              fd,
		UpdateTime:      time.Now(),
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		ReadBuf:         make(chan Message, 1024),
		WriteBuf:        make(chan Message, 1024),
		Server:          server,
	}
}

func (c *Conn) GetFd() int {
	return c.fd
}

func (c *Conn) Read() {
	var buf [1024]byte
	nbytes, _ := syscall.Read(c.fd, buf[:])
	if nbytes > 0 {
		//查询状态
		msgtype := int((buf[0] << 4) >> 4)
		//查询掩码
		mask := buf[1] >> 7

		if mask != 1 { //如果没有掩码，就直接将数据抛弃掉
			return
		}
		//sendStr = append(sendStr, buf[0])
		//查询数据的长度
		datalen := int64((buf[1] << 1) >> 1)
		//掩码默认从第二字节开始
		maskStart := 2 //掩码开始的字节数
		if msgtype == CloseMessage {
			//获取关闭信息
			closeReason := c.getMessage(buf[:], datalen, maskStart)
			c.CloseCode = binary.BigEndian.Uint16(closeReason[:2])
			c.CloseReason = closeReason[2:]
			c.Server.handler(c.fd, CONN_CLOSE)
			return
		}
		msg := &Message{}
		msg.MessageType = msgtype
		//如果data长度小于125，就直接取字节长度
		if datalen == 126 { //如果data长度等于126，就往后增加两个字节长度
			datalen = int64(binary.BigEndian.Uint16([]byte{buf[2], buf[3]}))
			maskStart = 4
		} else if datalen == 127 {
			var numBytes = []byte{}
			for i := 0; i < 8; i++ {
				numBytes = append(numBytes, buf[i+2])
			}
			datalen = int64(binary.BigEndian.Uint64(numBytes))
			maskStart = 10
		}
		msg.Content = c.getMessage(buf[:], datalen, maskStart)
		//发送内容
		c.UpdateTime = time.Now()
		msg.Conn = c
		c.Server.readMessageChan <- msg
		return
	}

}

func (c *Conn) getMessage(buf []byte, datalen int64, maskStart int) (content []byte) {
	maskSlice := make([]byte, 0) //mask
	for i := 0; i < 4; i++ {
		maskSlice = append(maskSlice, buf[i+maskStart])
	}
	datastart := int64(maskStart + 4)
	fmt.Printf("datastart======%d\n", datastart)
	maskIndex := 0
	content = make([]byte, 0)
	for i := datastart; i < datalen+datastart; i++ {
		content = append(content, buf[i]^maskSlice[maskIndex])
		if maskIndex == 3 {
			maskIndex = 0
		} else {
			maskIndex++
		}
	}
	return
}

func (c *Conn) Write(message []byte) (n int, err error) {
	msg := []byte{}
	msg = append(msg, 129)
	length := len(message)
	if length <= 125 {
		msg = append(msg, byte(length))
	} else if length > 125 && length < 65535 {
		msg = append(msg, 126)
		tmp := int16(length)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		msg = append(msg, bytesBuffer.Bytes()...)
	}
	msg = append(msg, message...)
	return syscall.Write(c.fd, msg)
}
