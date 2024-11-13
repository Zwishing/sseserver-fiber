package sseserver

import (
	"sync"

	"github.com/gofiber/fiber/v2"
)

// 全局变量声明
var (
	sseServer *Server   // SSE服务器实例
	once      sync.Once // 确保单例模式的并发安全
)

// Server 结构体定义了SSE服务器的基本结构
type Server struct {
	Broadcast chan<- SSEMessage // 广播通道,只写
	hub       *hub              // 内部连接集线器
}

// NewServer 创建并初始化一个新的SSE服务器实例
func newServer() *Server {
	s := Server{
		hub: newHub(), // 初始化内部hub
	}

	// 启动内部连接hub
	// hub作为Server的私有成员维护所有连接
	s.hub.Start()

	// 将hub的广播通道暴露为Server的公开只写通道
	s.Broadcast = s.hub.broadcast

	return &s
}

func Subscribe(ctx *fiber.Ctx, namespace string) error {
	once.Do(func() {
		sseServer = newServer()
	})
	return connect(ctx, sseServer.hub, namespace)
}

func SendSseMessage(msg SSEMessage) {
	sseServer.Broadcast <- msg
}

func Close() {
	if sseServer != nil && sseServer.hub != nil {
		close(sseServer.hub.broadcast)
		sseServer.hub.Shutdown()
	}
}
