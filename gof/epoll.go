package gof

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	EpollListener = syscall.EPOLLIN | syscall.EPOLLPRI | syscall.EPOLLERR | syscall.EPOLLHUP | unix.EPOLLET
	EpollRead     = syscall.EPOLLIN | syscall.EPOLLPRI | syscall.EPOLLERR | syscall.EPOLLHUP | unix.EPOLLET
)

type EpollObj struct {
	socket int    //socket连接
	epId   int    //epoll 创建的唯一描述符
	ip     string //socket监听的地址
	port   int    //socket监听的端口
}

//初始化epoll 包含创建socket,监听端口，以及创建epoll监听
func InitEpoll(ip string, port int) *EpollObj {
	ep := new(EpollObj)
	ep.ip = ip
	ep.port = port
	return ep.getScoket().listen().getGlobalFd()
}

//创建socket对象
func (e *EpollObj) getScoket() *EpollObj {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		fmt.Println("getScoket err:", err)
		os.Exit(1)
	}
	e.socket = fd
	return e
}

//监听端口
func (e *EpollObj) listen() *EpollObj {
	if err := syscall.SetNonblock(e.socket, true); err != nil {
		fmt.Println("setnonblock err", err)
		os.Exit(1)
	}
	//监听
	addr := syscall.SockaddrInet4{Port: e.port}
	ip := "0.0.0.0"
	if e.ip != "" {
		ip = e.ip
	}
	copy(addr.Addr[:], net.ParseIP(ip).To4())
	if err := syscall.Bind(e.socket, &addr); err != nil {
		fmt.Println("bind err:", err.Error())
		os.Exit(1)
	}
	if err := syscall.Listen(e.socket, 10); err != nil {
		fmt.Println("listen err:", err.Error())
		os.Exit(1)
	}
	return e
}

//创建epollfd对象，并加入监听
func (e *EpollObj) getGlobalFd() *EpollObj {
	//创建epfd
	epfd, err := syscall.EpollCreate1(0)
	log.Printf("getGlobalFd 创建的epfd为：%+v,e.fd:%d\n", epfd, e.socket)
	if err != nil {
		fmt.Println("epoll_create1 err:", err)
		os.Exit(1)
	}
	e.epId = epfd
	e.EpollCtl(e.socket, syscall.EPOLL_CTL_ADD)
	return e
}

//EpollCtl方法，添加、删除监听的fd
//fd 需要监听的fd对象
//status syscall.EPOLL_CTL_ADD添加  syscall.EPOLL_CTL_DEL删除
func (e *EpollObj) EpollCtl(fd, status int) {
	//通过EpollCtl将epfd加入到Epoll中，去监听
	if err := syscall.EpollCtl(e.epId, status, fd, &syscall.EpollEvent{Events: EpollListener, Fd: int32(fd)}); err != nil {
		fmt.Println("epoll_ctl err:", err)
		os.Exit(1)
	}
}

func (e *EpollObj) EpollWait(handle func(fd int, connType ConnStatus)) error {
	events := make([]syscall.EpollEvent, 3)
	n, err := syscall.EpollWait(e.epId, events, -1)
	if err != nil {
		fmt.Println("epoll_wait err:", e)
		return err
	}
	for i := 0; i < n; i++ {
		//如果是系统描述符，就建立一个新的连接
		connType := CONN_MESSAGE //默认是读内容
		if int(events[i].Fd) == e.socket {
			connType = CONN_NEW
		}
		handle(int(events[i].Fd), connType)
	}
	return nil
}

// func (e *EpollObj) Run() {
// 	/*第一个参数 domain
// 		// syscall.AF_INET，表示服务器之间的网络通信
// 		// syscall.AF_UNIX表示同一台机器上的进程通信
// 		// syscall.AF_INET6表示以IPv6的方式进行服务器之间的网络通信
// 	*/
// 	/*第二个参数 typ
// 	// syscall.SOCK_RAW，表示使用原始套接字，可以构建传输层的协议头部，启用IP_HDRINCL的话，IP层的协议头部也可以构造，就是上面区分的传输层socket和网络层socket。
// 	// syscall.SOCK_STREAM, 基于TCP的socket通信，应用层socket。
// 	// syscall.SOCK_DGRAM, 基于UDP的socket通信，应用层socket。
// 	*/
// 	/* 第三个参数 proto
// 	// IPPROTO_TCP 接收TCP协议的数据
// 	// IPPROTO_IP 接收任何的IP数据包
// 	// IPPROTO_UDP 接收UDP协议的数据
// 	// IPPROTO_ICMP 接收ICMP协议的数据
// 	// IPPROTO_RAW 只能用来发送IP数据包，不能接收数据。
// 	*/
// 	//创建系统描述符
// 	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// 	defer syscall.Close(fd)
// 	//将系统描述符设置为非阻塞
// 	if err = syscall.SetNonblock(fd, true); err != nil {
// 		fmt.Println("setnonblock err", err)
// 		os.Exit(1)
// 	}
// 	//监听
// 	addr := syscall.SockaddrInet4{Port: 8000}
// 	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())

// 	syscall.Bind(fd, &addr)
// 	syscall.Listen(fd, 10)
// 	//创建epfd
// 	epfd, e := syscall.EpollCreate1(0)
// 	if e != nil {
// 		fmt.Println("epoll_create1 err", err)
// 		os.Exit(1)
// 	}
// 	defer syscall.Close(epfd)
// 	s.event.Events = syscall.EPOLLIN
// 	s.event.Fd = int32(fd)
// 	//通过EpollCtl将epfd加入到Epoll中，去监听
// 	if e = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &s.event); e != nil {
// 		fmt.Println("epoll_ctl err:", e)
// 		os.Exit(1)
// 	}
// 	for {
// 		nevents, e := syscall.EpollWait(epfd, s.events[:], -1)
// 		fmt.Printf("%+v,%+v,%+v\n", nevents, epfd, fd)
// 		if e != nil {
// 			fmt.Println("epoll_wait err:", e)
// 			break
// 		}
// 		for ev := 0; ev < nevents; ev++ {
// 			//如果是系统描述符，就建立一个新的连接
// 			if int(s.events[ev].Fd) == fd {
// 				_, err := s.addConn((fd))
// 				if err != nil {
// 					fmt.Println("new conn err:", err.Error())
// 				}
// 			} else {
// 				//如果不是系统描述符，就通过Handler方法去处理fd的内容
// 				go s.handler(int(s.events[ev].Fd))
// 			}
// 		}
// 	}
// }
