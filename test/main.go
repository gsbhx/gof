package main

import (
	"GoPoll/gof"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
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
func (Ws) OnClose(c *gof.Conn) {
	fmt.Println("close:", c.GetFd())
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
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
