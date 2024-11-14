The core of repository is from [sseserver](https://github.com/mroth/sseserver), a server-sent events implementation in Go.
This is a version designed specifically for the Fiber web framework.

## Installation

```bash
go get github.com/zwishing/sseserver-fiber
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zwishing/sseserver-fiber"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	defer sseserver.Close()
	//CORS for external resources
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Cache-Control",
	}))

	app.Get("/sse", func(ctx *fiber.Ctx) error {
		err := sseserver.Subscribe(ctx,"sse")
		if err != nil {
			return err
		}
		go func() {
			// 使用time.Tick控制发送频率
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for i := 1; i <= 100; i++ {
				<-ticker.C // 等待下一个tick
				 sseserver.SendSseMessage(sseserver.SSEMessage{
					Event:     "processing-percent",
					Data:      []byte(fmt.Sprintf("%d%%", i)),
					Namespace: "sse",
				})
			}
			
		}()
		return nil
	})
	
	app.Listen(":8080")
}
```