# gof

### gof是什么

gof是一个开箱即用的websocket框架，通过golang的syscall函数直接调用linux的epoll模型，相比于gorilla/websocket框架，gof直接监听epoll句柄，因此性能更高。

### gof有什么

暂支持文本类型、二进制类型的内容接收以及文本类型的内容发送。

可配置连接超时时间。

可配置接收和发送消息的大小。

可自定义是否开启压缩模式。


### gof如何用

1、需要实现 gof/server.go 文件中的 WebSocketInterface接口，其中包含三个函数：
```
type WebSocketInterface interface {
    OnConnect(c *Conn) //握手完成之后的回调
    OnMessage(c *Conn, bytes []byte) //新消息回调
    OnClose(c *Conn, code uint16, reason []byte) //连接关闭时的回调
}
```
2、初始化一个server,然后执行server的run方法
```
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
    
    
    //现有的配置项支持三个,如果不配置的话，直接传nil就可以
    configure:=&gof.Conf{
		ReadBufferSize:    1024, //读取消息的缓冲区大小(byte)
		WriteBufferSize:   1024, //写入消息的缓冲区大小(byte)
		ConnectionTimeOut: 5,    //连接超时时间（秒）
		IsCompressOn: true, //是否开启压缩模式
		CompressLevel: 9,  //压缩等级
	}
	
func main(){
	//serve := gof.InitServer("0.0.0.0", 8801,Ws{},nil)
	serve := gof.InitServer("0.0.0.0", 8801,Ws{},configure)
    serve.Run()
}
```
