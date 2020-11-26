package gof

type ConnStatus int

const (
	CONN_NEW     ConnStatus = 1 //新连接
	CONN_CLOSE   ConnStatus = 2 //关闭连接
	CONN_MESSAGE ConnStatus = 3 //处理消息
)
