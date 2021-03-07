package gof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"time"
)

type Conn struct {
	s           *Server
	fd          int          //当前连接的文件描述符 fd
	updateTime  int64        //最新的更新时间，判断超时用
	handShake   chan Message //用于前期的验证和握手请求
	method      string       //请求方式 websocket必须是get请求方式
	closeCode   uint16       //关闭状态码
	closeReason []byte       //关闭原因
	canCompress bool         //是否支持压缩
}

func newConn(fd int, server *Server) *Conn {
	return &Conn{
		s:          server,
		fd:         fd,
		handShake:  make(chan Message, 1024),
		updateTime: time.Now().Unix(),
	}
}

func (c *Conn) GetFd() int {
	return c.fd
}

func (c *Conn) Read() {
	buf := c.s.readBufPool.Get().([]byte)
	defer func() {
		buf = make([]byte, c.s.readBufferSize)
		c.s.readBufPool.Put(buf)
	}()
	nbytes, _ := syscall.Read(c.fd, buf)
	if nbytes > 0 {
		fmt.Printf("%b\n",buf)
		Log.Info("Conn Read received message header:%b", buf[:2])
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
			closeReason := c.s.bytePool.Get().([]byte)
			closeReason = c.getMessage(buf[:])
			c.closeCode = binary.BigEndian.Uint16(closeReason[:2])
			c.closeReason = closeReason[2:]
			closeReason = []byte{}
			c.s.bytePool.Put(closeReason)
			c.s.closeChan <- c

		case BinaryMessage, TextMessage: //如果是二进制或者文本消息
			msg := c.s.messagePool.Get().(*Message)
			msg.MessageType = msgtype
			msg.Content = c.getMessage(buf[:])
			//发送内容
			newTime := time.Now().Unix()
			_ = c.s.checkTimeOutTree.Set(c.updateTime, newTime, c.fd)
			c.updateTime = newTime
			if c.canCompress == true && c.s.isComporessOn == true {
				// 一个缓存区压缩的内容
				var err error
				fmt.Printf("msg.Content===%b\n",msg.Content)
				msg.Content, err = DeCompress(msg.Content,c)
				fmt.Println("msg.Content===",msg.Content)
				if err != nil {
					Log.Fatal("解压句柄为 %d 的消息失败：%+v", c.fd, err.Error())
					return
				}
			}
			msg.Conn = c
			c.s.readMessageChan <- msg
			msg = &Message{
				Content: make([]byte, 0,c.s.writeBufferSize),
			}
			c.s.messagePool.Put(msg)
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
	maskSlice := c.s.bytePool.Get().([]byte)

	maskSlice = buf[maskStart : maskStart+4]
	datastart := int64(maskStart + 4)
	maskIndex := 0
	content := c.s.bytePool.Get().([]byte)
	for i := datastart; i < datalen+datastart; i++ {
		content = append(content, buf[i]^maskSlice[maskIndex])
		if maskIndex == 3 {
			maskIndex = 0
		} else {
			maskIndex++
		}
	}
	defer func() {
		maskSlice = make([]byte,0,c.s.writeBufferSize)
		c.s.bytePool.Put(maskSlice)
		content = make([]byte,0,c.s.writeBufferSize)
		c.s.bytePool.Put(content)

	}()
	return content
}

// @Author WangKan
// @Description //向句柄中写入文本内容
// @Date 2021/2/22 14:01
// @Param
// @return
func (c *Conn) Write(message []byte) {
	fmt.Println("message==",message)
	msg := c.s.bytePool.Get().([]byte)
	msg = append(msg, 129)

	if c.canCompress == true && c.s.isComporessOn == true {
	var err error
		msg[0] += 64
		fmt.Println("msg[0]=",msg[0])
		message, err = Compress(message, c.s.compressLevel)
		if err != nil {
			Log.Fatal("压缩 %d 的消息错误：%+v", c.fd, err)
		}
		message= append(message, 0)
		message=message[:len(message)-5]
	}
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
	_, _ = syscall.Write(c.fd, msg)
	msg = make([]byte, 0, c.s.writeBufferSize)
	c.s.bytePool.Put(msg)
}
