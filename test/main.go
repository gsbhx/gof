package main

import (
	"gof"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)


type Ws struct {
}

func (Ws) OnConnect(c *gof.Conn) {
	fmt.Println("connect:", c.GetFd())
}
func (Ws) OnMessage(c *gof.Conn, bytes []byte) {
	fmt.Println("read:", string(bytes))
	c.Write(bytes)
}
func (Ws) OnClose(c *gof.Conn,code uint16, reason []byte) {
	fmt.Println("close:", c.GetFd(),"closeCode:",code," closeReason:",string(reason))
}


func main() {
	serve := gof.InitServer("0.0.0.0", 8801,Ws{})
	go func() {
		serve.Run()
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			serve.Close()
			//time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
