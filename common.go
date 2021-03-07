package gof

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
	TextMessage   = 1  //文本消息
	BinaryMessage = 2  //二进制消息
	CloseMessage  = 8  //关闭消息
	PingMessage   = 9  //ping消息
	PongMessage   = 10 //pong消息
)

type Message struct {
	Conn        *Conn
	MessageType int
	Content     []byte
}

type Conf struct {
	ReadBufferSize    int
	WriteBufferSize   int
	ConnectionTimeOut int64
	CompressLevel     int
	IsCompressOn      bool
}
