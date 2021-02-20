package gof

import (
	"net/http"
	"strings"
)

// HandshakeError describes an error with the handshake from the peer.
type HandshakeError struct {
	message string
}

func (e HandshakeError) Error() string { return e.message }

type Upgrader struct {
	// Error specifies the function for generating HTTP error responses. If Error
	// is nil, then http.Error is used to generate the HTTP response.
	Error func(status int, reason error)

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin func(header http.Header) bool

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this field is not nil, then the Upgrade method negotiates a
	// subprotocol by selecting the first match in this list with a protocol
	// requested by the client. If there's no match, then no protocol is
	// negotiated (the Sec-Websocket-Protocol header is not included in the
	// handshake response).
	Subprotocols []string

	// EnableCompression specify if the server should attempt to negotiate per
	// message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool
}

func (u *Upgrader) returnError(status int, reason string) (*Conn, error) {
	err := HandshakeError{reason}
	return nil, err
}

func (u *Upgrader) Upgrade(fd int, header map[string]string,s *Server) (*Conn, error) {
	const badHandshake = "websocket: the client is not using the websocket protocol: "

	if header["Connection"]!="Upgrade" && header["Connection"]!="upgrade"{
		return u.returnError(http.StatusBadRequest, badHandshake+"'upgrade' token not found in 'Connection' header")
	}

	if header["Upgrade"]!="websocket" {
		return u.returnError(http.StatusBadRequest, badHandshake+"'websocket' token not found in 'Upgrade' header")
	}

	if header["Method"] != "GET" {
		return u.returnError(http.StatusMethodNotAllowed, badHandshake+"request method is not GET")
	}
	if  header["Sec-WebSocket-Version"] != "13"   {
		return u.returnError(http.StatusBadRequest, "websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header")
	}

	//if _, ok := header["Sec-WebSocket-Extensions"]; ok {
	//	return u.returnError(http.StatusInternalServerError, "websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
	//}


	challengeKey := header["Sec-WebSocket-Key"]
	if challengeKey == "" {
		return u.returnError(http.StatusBadRequest, "websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank")
	}
	c := newConn(fd,s)
	// Use larger of hijacked buffer and connection write buffer for header.
	wf:=[]byte{}
	wf = append(wf, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	wf = append(wf, computeAcceptKey(challengeKey)...)
	wf = append(wf, "\r\n"...)

	wf = append(wf, "\r\n"...)
	c.handShake <-Message{
		MessageType: -1,
		Content:     wf,
	}
	return c, nil
}


func (u *Upgrader) selectSubprotocol(r *http.Request, responseHeader http.Header) string {
	if u.Subprotocols != nil {
		clientProtocols := Subprotocols(r)
		for _, serverProtocol := range u.Subprotocols {
			for _, clientProtocol := range clientProtocols {
				if clientProtocol == serverProtocol {
					return clientProtocol
				}
			}
		}
	} else if responseHeader != nil {
		return responseHeader.Get("Sec-Websocket-Protocol")
	}
	return ""
}

// Subprotocols returns the subprotocols requested by the client in the
// Sec-Websocket-Protocol header.
func Subprotocols(r *http.Request) []string {
	h := strings.TrimSpace(r.Header.Get("Sec-Websocket-Protocol"))
	if h == "" {
		return nil
	}
	protocols := strings.Split(h, ",")
	for i := range protocols {
		protocols[i] = strings.TrimSpace(protocols[i])
	}
	return protocols
}

// writeHook is an io.Writer that records the last slice passed to it vio
// io.Writer.Write.
type writeHook struct {
	p []byte
}

func (wh *writeHook) Write(p []byte) (int, error) {
	wh.p = p
	return len(p), nil
}
