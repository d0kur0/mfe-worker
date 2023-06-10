package src

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type HttpServer struct {
	di *DIContainer
}

func (h *HttpServer) SetupHttpHandlers() error {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		//h.di.queue.AddToQueue("1635", "test")

		return c.String(http.StatusOK, "Hello, World!")
	})

	log.Printf("Server started at http://localhost:%d", h.di.configMap.HttpPort)
	return e.Start(fmt.Sprintf(":%d", h.di.configMap.HttpPort))
}

func NewHttpServer(di *DIContainer) (*HttpServer, error) {
	return &HttpServer{
		di: di,
	}, nil
}
