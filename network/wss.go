package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/liaodiansm/mtg/essentials"
	utls "github.com/refraction-networking/utls"
)

type wssDialer struct {
	Dialer

	password     []byte
	proxyAddress string
}

// WSConn 封装 websocket.Conn
type WSConn struct {
	conn *websocket.Conn
	rbuf []byte
}

// NewWSConn 包装已有 websocket.Conn
func NewWSConn(c *websocket.Conn) *WSConn {
	return &WSConn{conn: c}
}

// --- io.Reader ---
func (w *WSConn) Read(p []byte) (n int, err error) {
	if len(w.rbuf) > 0 {
		n = copy(p, w.rbuf)
		w.rbuf = w.rbuf[n:]
		return n, nil
	}

	for {
		mt, message, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		if mt != websocket.BinaryMessage {
			// 忽略非二进制消息
			continue
		}
		w.rbuf = message
		n = copy(p, w.rbuf)
		w.rbuf = w.rbuf[n:]
		return n, nil
	}
}

// --- io.Writer ---
func (w *WSConn) Write(p []byte) (n int, err error) {
	err = w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// --- CloseableReader / CloseableWriter ---
func (w *WSConn) CloseRead() error {
	return w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (w *WSConn) CloseWrite() error {
	return w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// --- net.Conn 方法 ---
func (w *WSConn) Close() error {
	return w.conn.Close()
}

func (w *WSConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WSConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WSConn) SetDeadline(t time.Time) error {
	if err := w.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return w.conn.SetWriteDeadline(t)
}

func (w *WSConn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WSConn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

func (s wssDialer) Dial(network, address string) (essentials.Conn, error) {
	return s.DialContext(context.Background(), network, address)
}

// 自定义 TLS 指纹
func utlsDialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {

	// 尝试去除端口，获取 SNI
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	// 标准 Dialer 建立 TCP 连接
	var d net.Dialer
	tcpConn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	// 创建 uTLS 客户端连接（模拟浏览器）
	uconn := utls.UClient(tcpConn, &utls.Config{ServerName: host}, utls.Hello360_7_5)

	// 设置超时，避免一直等待
	if dl, ok := ctx.Deadline(); ok {
		_ = uconn.SetDeadline(dl)
	}

	// tls 握手
	if err := uconn.Handshake(); err != nil {
		//Debug(err)
		tcpConn.Close()
		return nil, err
	}

	// 握手后清除超时设置
	_ = uconn.SetDeadline(time.Time{})

	return uconn, nil
}

func (s wssDialer) DialContext(ctx context.Context, network, address string) (essentials.Conn, error) {

	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("%s network type is not supported", network)
	}

	header := make(http.Header)
	header.Set("X-Password", string(s.password))
	header.Set("X-Target", address)

	// conn, _, err := websocket.DefaultDialer.Dial("wss://"+s.proxyAddress, header)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot dial to the websocket: %w", err)
	// }

	dialer := websocket.Dialer{
		NetDialTLSContext: utlsDialTLSContext,
		HandshakeTimeout:  30 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial("wss://"+s.proxyAddress, header)
	if err != nil {
		return nil, fmt.Errorf("cannot dial to the websocket: %w", err)
	}

	return NewWSConn(conn), nil
}

// NewWssDialer build a new dialer from a given one (so, in theory you can
// chain here). Proxy parameters are passed with URI in a form of:
//
//	wss://[:[password]]@host:port
func NewWssDialer(baseDialer Dialer, proxyURL *url.URL) (Dialer, error) {
	if _, _, err := net.SplitHostPort(proxyURL.Host); err != nil {
		return nil, fmt.Errorf("incorrect url %s", proxyURL.Redacted())
	}

	dialer := wssDialer{
		Dialer:       baseDialer,
		proxyAddress: proxyURL.Host,
	}

	if proxyURL.User != nil {
		password, isSet := proxyURL.User.Password()
		if isSet {
			dialer.password = []byte(password)
		}
	}

	return dialer, nil
}
