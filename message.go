package sseserver 

// SSEMessage 用于在Server-Sent Event流上发送的消息结构体
//
// 注意: Namespace 不属于SSE规范的一部分,仅在内部用于
// 将消息映射到适当的HTTP虚拟端点
type SSEMessage struct {
	Event     string // 消息的事件作用域 [可选]
	Data      []byte // 消息载荷
	Namespace string // 消息的命名空间,用于匹配客户端订阅
}

// sseFormat 将SSE消息格式化为准备发送的字节串
// 通过预分配合适大小的buffer来优化性能
func (msg SSEMessage) sseFormat() []byte {
	// 预分配buffer容量 = "event:" + "data:" + Event内容 + Data内容 + 换行符
	b := make([]byte, 0, 6+5+len(msg.Event)+len(msg.Data)+3)

	// 如果Event不为空,添加event字段
	if msg.Event != "" {
		b = append(b, "event:"...)
		b = append(b, msg.Event...)
		b = append(b, '\n')
	}

	// 添加data字段和结束换行符
	b = append(b, "data:"...)
	b = append(b, msg.Data...)
	b = append(b, '\n', '\n')
	return b
}
