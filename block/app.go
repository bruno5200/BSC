package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	a "github.com/bruno5200/CSM/block/application"
	h "github.com/bruno5200/CSM/block/infrastructure/handler"
	r "github.com/bruno5200/CSM/block/infrastructure/repository"
	"github.com/bruno5200/CSM/block/router"
	db "github.com/bruno5200/CSM/database"
	"github.com/bruno5200/CSM/env"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var e = env.Env()

func init() { env.Init() }

func main() {

	go db.NewPostgresDB()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// CORS for external resources
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
		AllowHeaders: "Cache-Control, Accept, Content-Type, Content-Length, Accept-Encoding, Authorization",
	}))

	app.Use(recover.New())
	app.Use(favicon.New())
	app.Use(logger.New())
	app.Use(requestid.New())

	app.Get("/metrics", monitor.New(monitor.Config{
		Title:   "IDC API Metrics",
		Refresh: 2 * time.Second,
	}))

	router.BlockRouter(app, h.NewBlockHandler(a.NewBlockService(r.NewBlockRepository(db.PostgresDB()))))

	app.Get("/*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString(`<h1>404 No Encontrado</h1>`)
	})

	port := ":" + e.GetPort()

	if e.GetSecure() {
		go func() {
			log.Printf(`Running with TLS in https://localhost%v`, port)
			if err := app.ListenTLS(
				port,
				"./certs/storage.cert",
				"./certs/storage.key",
			); err != nil {
				log.Panic(err)
			}
		}()
	} else {
		go func() {
			log.Printf(`Running in http://localhost%v`, port)
			if err := app.Listen(port); err != nil {
				log.Panic(err)
			}
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	_ = app.Shutdown()

	log.Println("Running cleanup tasks...")
	//database.CloseDb()
	db.Close()
	// redisConn.Close()
	log.Println("Fiber was successful shutdown.")
}
