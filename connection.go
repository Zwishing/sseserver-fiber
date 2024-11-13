// sseserver 包实现了基于SSE(Server-Sent Events)的服务器推送功能
package sseserver

import (
	"bufio"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"sync/atomic"
	"time"
)

const (
	// 连接缓冲区大小
	connBufSize = 256
	// keepalive 心跳间隔时间
	keepAliveInterval = 15 * time.Second
	// HTTP 响应状态码
	httpOK = 200
)

// connection 结构体表示单个SSE连接
type connection struct {
	ctx       *fiber.Ctx    // HTTP上下文
	created   time.Time     // 连接创建时间
	send      chan []byte   // 消息发送通道
	namespace string        // 命名空间,用于消息分组
	msgsSent  uint64        // 已发送消息计数
	done      chan struct{} // 连接关闭信号
}

// newConnection 创建新的SSE连接
func newConnection(ctx *fiber.Ctx, namespace string) *connection {
	return &connection{
		ctx:       ctx,
		send:      make(chan []byte, connBufSize),
		created:   time.Now(),
		namespace: namespace,
		done:      make(chan struct{}),
	}
}

// connectionStatus 结构体用于输出连接状态信息
type connectionStatus struct {
	Path      string `json:"request_path"` // 请求路径
	Namespace string `json:"namespace"`    // 命名空间
	Created   int64  `json:"created_at"`   // 创建时间戳
	ClientIP  string `json:"client_ip"`    // 客户端IP
	UserAgent string `json:"user_agent"`   // 用户代理
	MsgsSent  uint64 `json:"msgs_sent"`    // 已发送消息数
}

// Status 返回当前连接状态
func (c *connection) Status() connectionStatus {
	return connectionStatus{
		Path:      c.ctx.Path(),
		Namespace: c.namespace,
		Created:   c.created.Unix(),
		ClientIP:  c.ctx.IP(),
		UserAgent: c.ctx.Get("User-Agent"),
		MsgsSent:  atomic.LoadUint64(&c.msgsSent),
	}
}

// write 包装写入操作,统一错误处理
func write(w *bufio.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return err
	}
	return w.Flush()
}

// writer 处理消息写入和心跳保活
func (c *connection) writer(h *hub) {
	keepaliveTickler := time.NewTicker(keepAliveInterval)
	keepaliveMsg := []byte(":keepalive\n")
	defer keepaliveTickler.Stop()

	c.ctx.Status(httpOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case msg, ok := <-c.send:
				if !ok {
					return
				}
				if err := write(w, msg); err != nil {
					h.unregister <- c
					return
				}
				atomic.AddUint64(&c.msgsSent, 1)

			case <-keepaliveTickler.C:
				if err := write(w, keepaliveMsg); err != nil {
					h.unregister <- c
					return
				}
			}
		}
	}))
}

// connectHandler 创建处理SSE连接的HTTP处理器
func connectHandler(h *hub, namespace string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// namespace := strings.TrimPrefix(c.Path(), "/")
		// if namespace == "" {
		// 	namespace = "/"
		// } else {
		// 	namespace = "/" + namespace
		// }

		conn := newConnection(c, namespace)
		h.register <- conn
		conn.writer(h)
		return nil
	}
}

func connect(c *fiber.Ctx, h *hub, namespace string) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	conn := newConnection(c, namespace)
	h.register <- conn
	conn.writer(h)
	return nil
}
