package gof

import (
	"compress/flate"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// Handler Server 注册接口
type WebSocketInterface interface {
	OnConnect(c *Conn)                           //握手完成之后的回调
	OnMessage(c *Conn, bytes []byte)             //新消息回调
	OnClose(c *Conn, code uint16, reason []byte) //连接关闭时的回调
}

type Server struct {
	ep                *EpollObj
	conns             sync.Map //当前的所有连接
	checkTimeOutTree  *AVLTree
	receiveFdBytes    chan *Conn
	handle            WebSocketInterface
	readMessageChan   chan *Message
	closeChan         chan *Conn //需要关闭的所有Conn
	readBufferSize    int
	writeBufferSize   int
	connectionTimeout int64
	bytePool          *sync.Pool //[]byte 的池子
	readBufPool       *sync.Pool // [1024]byte的池子，用于接收fd描述符上的内容
	messagePool       *sync.Pool //Message的池子，用于接收消息并返给服务端
	isComporessOn     bool
	compressLevel     int
}

func InitServer(ip string, port int, handle WebSocketInterface, conf *Conf) *Server {
	ep := InitEpoll(ip, port)
	serv := &Server{
		ep:                ep,
		receiveFdBytes:    make(chan *Conn, 1024),
		readMessageChan:   make(chan *Message, 1024),
		handle:            handle,
		conns:             sync.Map{},
		checkTimeOutTree:  NewAvlTree(),
		closeChan:         make(chan *Conn, 1024),
		readBufferSize:    1024,
		writeBufferSize:   1024,
		connectionTimeout: 30,
		isComporessOn:     false,
		compressLevel:     0,
	}
	if conf != nil {
		if conf.ReadBufferSize > 0 {
			serv.readBufferSize = conf.ReadBufferSize
		}
		if conf.WriteBufferSize > 0 {
			serv.writeBufferSize = conf.WriteBufferSize
		}
		if conf.ConnectionTimeOut > 0 {
			serv.connectionTimeout = conf.ConnectionTimeOut
		}
		if conf.IsCompressOn == true {
			serv.isComporessOn = true
			serv.compressLevel = conf.CompressLevel
			if conf.CompressLevel==0{
				serv.compressLevel = flate.BestCompression
			}
		}
	}

	serv.readBufPool = &sync.Pool{New: func() interface{} { return make([]byte, serv.readBufferSize) }}
	serv.messagePool = &sync.Pool{New: func() interface{} {
		return &Message{
			Content: make([]byte,0,serv.writeBufferSize),
		}
	}}
	serv.bytePool = &sync.Pool{New: func() interface{} { return make([]byte, 0, serv.writeBufferSize) }}

	return serv
}

func (s *Server) Run() {
	s.checkTimeOut() //如果过期，就关闭conn
	s.checkMessage() //如果有消息，就调用 conn.read方法解包
	s.getMessage()   //如果有新的消息，就走消息处理的逻辑
	s.closeConn()
	s.EpollWait()
}

var upgrader = Upgrader{}

// @Author WangKan
// @Description //当wait方法取到内容后，会回调此方法，对fd进行处理
// @Date 2021/2/2 21:39
func (s *Server) handler(fd int, connType ConnStatus) {
	switch connType {
	case CONN_NEW:
		newFd := s.addConn(fd)
		//Upgrader to http header
		s.handShaker(newFd)
		//s.messageChan<-newFd
	case CONN_MESSAGE:
		Log.Info("接收到描述符为%v的消息", fd)
		c, ok := s.conns.Load(fd)
		if !ok {
			Log.Info("描述符fd 为 %d 的s.conns 不存在！", fd)
			return
		}
		s.receiveFdBytes <- c.(*Conn)
	default:
		panic("no connType")
	}
}

// @Author WangKan
// @Description //握手方法，接收conn的头信息，解析并向客户端返回response信息
// @Date 2021/2/2 21:38
func (s *Server) handShaker(fd int) {
	header, length := GetHeaderContent(fd)
	headerMap := FormatHeader(header, length)
	newConn, err := upgrader.Upgrade(fd, headerMap, s)
	if err != nil {
		Log.Error("upgrade err: %+v", err.Error())
		return
	}
	heade := <-newConn.handShake
	fmt.Println(string(heade.Content))
	n, err := syscall.Write(fd, heade.Content)
	Log.Info("send handshaker message n:%+v, err: %+v, fd:%d, newConn:%+v\n", n, err, fd, newConn)

	if err != nil {
		Log.Error("send handshaker message err: %+v,fd:%d,%+v", err.Error(), fd, newConn)
		return
	}
	s.handle.OnConnect(newConn)
	Log.Info("要加入到链接库中的fd:%v", fd)
	s.conns.Store(fd, newConn)
	s.checkTimeOutTree.Add(newConn.updateTime, fd)
}

// @Author WangKan
// @Description //如果有新的连接，就取出系统中的fd，添加到当前的conns中。
// @Date 2021/2/2 21:37
func (s *Server) addConn(fd int) (newFd int) {
	newFd, _, err := syscall.Accept(fd)
	fmt.Printf("系统描述符新建的链接：%+v\n", newFd)
	if err != nil {
		fmt.Println("accept err: ", err)
		return
	}
	//设置fd为非阻塞
	if err := syscall.SetNonblock(newFd, true); err != nil {
		os.Exit(1)
	}
	//把这个链接加入到epoll中
	s.ep.eAdd(newFd)
	return
}

// @Author WangKan
// @Description //wait方法 阻塞式，当epoll中有数据的时候就取出数据并进行处理
// @Date 2021/2/2 21:37
func (s *Server) EpollWait() {
	for {
		err := s.ep.eWait(s.handler)
		if err != nil {
			Log.Error("epoll wait error: %s", err.Error())
			continue
		}
	}
}

// @Author WangKan
// @Description //如果有新的消息进来，就通过当前Conn的read方法去取message 并判断类型
// @Date 2021/2/2 18:12
func (s *Server) checkMessage() {
	go func() {
		for c := range s.receiveFdBytes {
			c.Read()
		}
	}()
}

// @Author WangKan
// @Description //如果有新的消息，就走消息处理的逻辑
// @Date 2021/2/2 18:13
func (s *Server) getMessage() {
	go func() {
		for c := range s.readMessageChan {
			content := c.Content
			s.handle.OnMessage(c.Conn, content)
		}
	}()
}

// @Author WangKan
// @Description //判断当前的s.closeChan中是否有数据，如果有就取出并删除，否则就一直阻塞
// @Date 2021/2/2 21:36
func (s *Server) closeConn() {
	go func() {
		for c := range s.closeChan {
			//先从conns中删掉当前的连接
			s.closeFd(c)
		}
	}()
}

// @Author WangKan
// @Description //判断conn是否已经超时，如果超时就关闭这个conn
// @Date 2021/2/2 21:35
func (s *Server) checkTimeOut() {
	go func() {
		for {
			//s.conns.Range(func(k, v interface{}) bool {
			//	if time.Now().Sub(v.(*Conn).updateTime) >= time.Second* s.connectionTimeout {
			//		Log.Info("fd 为 %d 的连接即将被断开\n", v.(*Conn).fd)
			//		s.closeFd(v.(*Conn))
			//	}
			//	return true
			//})
			//time.Sleep(time.Second * 2)
			/**********************************更改删除超时连接的结构为平衡二叉树*************************************/
			//给定一个值，获取小于该值的所有元素
			timeOutint64 := time.Now().Unix() - s.connectionTimeout
			//fmt.Println("当前时间：", time.Now().Unix())
			expiredKeys := s.checkTimeOutTree.GetLessThanKey(timeOutint64)
			//删除conns中的已超时的fd
			if expiredKeys != nil {
				for _, v := range expiredKeys {
					slice := make([]int, 0, 1024)
					for _, vv := range s.checkTimeOutTree.Get(v) {
						slice = append(slice, vv)
					}
					for i := 0; i < len(slice); i++ {
						c, _ := s.conns.Load(slice[i])
						s.closeFd(c.(*Conn))
					}

				}
			}
			//删除树中已超时的fd
			//s.checkTimeOutTree.RemoveOneNodeAndChilds(node.Key)

			//fmt.Println(s.checkTimeOutTree.InOrder(-1))

			time.Sleep(time.Second)
			/**********************************更改删除超时连接的结构为平衡二叉树END**********************************/
		}
	}()
}

// @Author WangKan
// @Description //关闭某一个fd, 从conns中删除 conn,从epoll实例中删除fd,并从系统中删除fd
// @Date 2021/2/2 21:46
// @Param [c] //Conn
func (s *Server) closeFd(c *Conn) {
	//从当前的epoll中删除fd
	s.ep.eDel(c.fd)
	//从系统中关闭当前fd
	_ = syscall.Close(c.fd)
	//从 s.conns中删除当前fd
	Log.Info("正在删除fd=%d的连接", c.fd)
	_ = s.checkTimeOutTree.RemoveNodeValue(c.updateTime, c.fd)
	s.conns.Delete(c.fd)
	s.handle.OnClose(c, c.closeCode, c.closeReason)
}

// @Author WangKan
// @Description //系统发送Ctrl+c信号的时候，调用此方法关闭所有的连接
// @Date 2021/2/2 21:40
func (s *Server) Close() {
	s.CloseFds()
	if err := syscall.Close(s.ep.epId); err != nil {
		Log.Error("Server Close epId err:%+v", err.Error())
	}

}

// @Author WangKan
// @Description //获取所有的conn 并调用关闭方法
// @Date 2021/2/2 21:34
func (s *Server) CloseFds() {
	s.conns.Range(func(k, v interface{}) bool {
		s.closeFd(v.(*Conn))
		return true
	})
}
