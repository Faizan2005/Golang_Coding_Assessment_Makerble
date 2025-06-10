package routes

import (
	"github.com/gofiber/fiber/v2"
)

type APIServer struct {
	listenAddr string
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	app := fiber.New()

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from POST route!")
	})

	app.Listen(s.listenAddr)
}
