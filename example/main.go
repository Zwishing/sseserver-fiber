package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"sseserver-fiber"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	//CORS for external resources
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Cache-Control",
	}))

	app.Get("/sse", func(ctx *fiber.Ctx) error {
		err := sseserver.RegisterConnection(ctx)
		if err != nil {
			return err
		}
		go func() {
			server := sseserver.GetSseServer()

			// 使用time.Tick控制发送频率
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for i := 1; i <= 100; i++ {
				<-ticker.C // 等待下一个tick
				server.Broadcast <- sseserver.SSEMessage{
					Event:     "processing-percent",
					Data:      []byte(fmt.Sprintf("%d%%", i)),
					Namespace: "/sse/process",
				}
			}
		}()
		return nil
	})

	app.Listen(":8080")
}
