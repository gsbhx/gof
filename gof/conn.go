package gof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"time"
)

type Conn struct {
	s           *Server
	fd          int          //当前连接的文件描述符 fd
	updateTime  time.Time    //最新的更新时间，判断超时用
	handShake   chan Message //用于前期的验证和握手请求
	method      string       //请求方式 websocket必须是get请求方式
	closeCode   uint16       //关闭状态码
	closeReason []byte       //关闭原因
	bytePool    *sync.Pool    //[]byte 的池子
	readBufPool *sync.Pool    // [1024]byte的池子，用于接收fd描述符上的内容
	messagePool *sync.Pool    //Message的池子，用于接收消息并返给服务端
}

func newConn(fd int, server *Server) *Conn {
	return &Conn{
		s:           server,
		fd:          fd,
		handShake:   make(chan Message, 1024),
		updateTime:  time.Now(),
		bytePool:    &sync.Pool{New: func() interface{} { return make([]byte, 0) }},
		readBufPool: &sync.Pool{New: func() interface{} { return [1024]byte{} }},
		messagePool: &sync.Pool{New: func() interface{} { return &Message{} }},
	}
}

func (c *Conn) GetFd() int {
	return c.fd
}

func (c *Conn) Read() {
	buf := c.readBufPool.Get().([1024]byte)
	defer func() {
		buf = [1024]byte{}
		c.readBufPool.Put(buf)
	}()
	nbytes, _ := syscall.Read(c.fd, buf[:])
	if nbytes > 0 {
		//查询状态
		msgtype := int((buf[0] << 4) >> 4)
		//查询掩码
		mask := buf[1] >> 7

		if mask != 1 { //如果没有掩码，就直接将数据抛弃掉
			return
		}
		//根据消息类型做处理
		switch msgtype {
		case CloseMessage:
			//获取关闭信息
			closeReason := c.bytePool.Get().([]byte)
			closeReason = c.getMessage(buf[:])
			c.closeCode = binary.BigEndian.Uint16(closeReason[:2])
			c.closeReason = closeReason[2:]
			closeReason = []byte{}
			c.bytePool.Put(closeReason)
			c.s.closeChan<-c

		case BinaryMessage, TextMessage: //如果是二进制或者文本消息
			msg := c.messagePool.Get().(*Message)
			msg.MessageType = msgtype
			msg.Content = c.getMessage(buf[:])
			//发送内容
			c.updateTime = time.Now()
			msg.Conn = c
			c.s.readMessageChan <- msg
			msg = &Message{}
			c.messagePool.Put(msg)
		}

		return
	}

}

// @Author WangKan
// @Description //将接收到的消息解包
// @Date 2021/2/2 18:14
// @Param
// @return
func (c *Conn) getMessage(buf []byte) []byte {
	//查询数据的长度
	datalen := int64((buf[1] << 1) >> 1)
	//掩码默认从第二字节开始
	maskStart := 2
	//如果data长度小于125，就直接取字节长度
	if datalen == 126 { //如果data长度等于126，就往后增加两个字节长度
		datalen = int64(binary.BigEndian.Uint16(buf[2:3]))
		maskStart = 4
	} else if datalen == 127 {
		datalen = int64(binary.BigEndian.Uint64(buf[2:9]))
		maskStart = 10
	}
	//4位的掩码
	maskSlice := c.bytePool.Get().([]byte)

	maskSlice = buf[maskStart : maskStart+4]
	datastart := int64(maskStart + 4)
	maskIndex := 0
	content := c.bytePool.Get().([]byte)
	for i := datastart; i < datalen+datastart; i++ {
		content = append(content, buf[i]^maskSlice[maskIndex])
		if maskIndex == 3 {
			maskIndex = 0
		} else {
			maskIndex++
		}
	}
	defer func() {
		maskSlice = []byte{}
		c.bytePool.Put(maskSlice)
		content = []byte{}
		c.bytePool.Put(content)

	}()
	return content
}

func (c *Conn) Write(message []byte) {
	msg := c.bytePool.Get().([]byte)
	msg = append(msg, 129)
	length := len(message)
	if length <= 125 {
		msg = append(msg, byte(length))
	} else if length > 125 && length < 65535 {
		msg = append(msg, 126)
		tmp := int16(length)
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		msg = append(msg, bytesBuffer.Bytes()...)
	}
	msg = append(msg, message...)
	fmt.Println(msg)
	_, _ = syscall.Write(c.fd, msg)
	msg = []byte{}
	c.bytePool.Put(msg)
}
