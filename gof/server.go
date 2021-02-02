package gof

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// Handler Server 注册接口
type WebSocketInterface interface {
	OnConnect(c *Conn)                           // OnConnect 当TCP长连接建立成功是回调
	OnMessage(c *Conn, bytes []byte)             // OnMessage 当客户端有数据写入是回调
	OnClose(c *Conn, code uint16, reason []byte) // OnClose 当客户端主动断开链接或者超时时回调,err返回关闭的原因
}

type Server struct {
	ep              *EpollObj
	conns           sync.Map //当前的所有连接
	receiveFdBytes  chan *Conn
	handle          WebSocketInterface
	readMessageChan chan *Message
}

func InitServer(ip string, port int, handle WebSocketInterface) *Server {
	ep := InitEpoll(ip, port)
	return &Server{
		ep:              ep,
		receiveFdBytes:  make(chan *Conn, 1024),
		readMessageChan: make(chan *Message,1024),
		handle:          handle,
		conns:           sync.Map{},
	}
}

func (s *Server) Run() {
	fmt.Printf("%+v\n", s.ep)
	s.checkTimeOut()
	s.checkMessage()
	s.getMessage()
	s.EpollWait()
}

func (s *Server) Close() {
	s.CloseFds()
	s.ep.EpollDel(s.ep.epId)
}

var upgrader = Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handler(fd int, connType ConnStatus) {
	switch connType {
	case CONN_NEW:
		newFd := s.addConn(fd)
		//Upgrader to http header
		s.handShaker(newFd)
		//s.messageChan<-newFd
	case CONN_MESSAGE:
		Log.Info("接收到描述符为%v的消息", fd)
		c, _ := s.conns.Load(fd)
		fmt.Println("接收到消息的conn为：", c.(*Conn))
		s.receiveFdBytes <- c.(*Conn)

	case CONN_CLOSE:
		c, _ := s.conns.Load(fd)
		c = c.(*Conn)
		s.handle.OnClose(c.(*Conn), c.(*Conn).CloseCode, c.(*Conn).CloseReason)
		s.closeFd(fd)
	default:
		panic("no connType")
	}
}

//
func (s *Server) handShaker(fd int) {
	header, length := MakeHeaderMap(fd)
	headerMap := FormatHeader(header, length)
	newConn, err := upgrader.Upgrade(fd, headerMap, s)
	if err != nil {
		fmt.Println("upgrade err:", err.Error())
	}
	heade := <-newConn.WriteBuf
	_, _ = syscall.Write(fd, heade.Content)
	s.handle.OnConnect(newConn)
	Log.Info("要加入到链接库中的fd:%v", fd)
	s.conns.Store(fd, newConn)
}

func (s *Server) closeFd(fd int) {
	//先从conns中删掉当前的连接
	Log.Info("要删除的fd:%v", fd)
	s.conns.Delete(fd)
	//从当前的epoll中删除fd
	s.ep.EpollDel(fd)
	//从系统中关闭当前fd
	_ = syscall.Close(fd)
}

func (s *Server) addConn(fd int) (newFd int) {
	newFd, _, err := syscall.Accept(fd)
	fmt.Printf("系统描述符新建的链接：%+v\n", newFd)
	if err != nil {
		fmt.Println("accept err: ", err)
		return
	}
	if err := syscall.SetNonblock(newFd, true); err != nil {
		fmt.Println("setnonblock err", err)
		os.Exit(1)
	}
	//把这个链接加入到epoll中
	s.ep.EpollAdd(newFd)
	return
}

func (s *Server) EpollWait() {
	for {
		err := s.ep.EpollWait(s.handler)
		if err != nil {
			Log.Error("epoll wait error: %s", err.Error())
			continue
		}
	}
}


func (s *Server) checkMessage() {
	go func() {
		for c := range s.receiveFdBytes {
			c.Read()
			//content := <-c.ReadBuf
			//s.handle.OnMessage(c, content.Content)
		}
	}()
}

func (s *Server) getMessage(){
	go func() {
		for c:=range s.readMessageChan{
			content :=c.Content
			s.handle.OnMessage(c.Conn, content)
		}
	}()
}

func (s *Server) checkTimeOut() {
	fmt.Println("开始checkTimeOut")
	go func() {
		for {
			s.conns.Range(func(k, v interface{}) bool {
				c := v.(*Conn)
				if time.Now().Sub(c.UpdateTime) >= time.Second*100 {
					Log.Info("fd 为 %d 的连接即将被断开\n", c.fd)
					s.handler(c.fd, CONN_CLOSE)
				}
				return true
			})
		}
	}()
}

func (s *Server) CloseFds() {
	for {
		s.conns.Range(func(k, v interface{}) bool {
			c := v.(*Conn)
			s.handler(c.fd, CONN_CLOSE)
			return true
		})
	}
}
