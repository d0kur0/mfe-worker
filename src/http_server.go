package src

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type HttpServer struct {
	pipeline  *Pipeline
	configMap *ConfigMap
}

func (ctx *HttpServer) SetupHttpHandlers() error {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e.Start(fmt.Sprintf(":%d", ctx.configMap.HttpPort))
}

func NewHttpServer(configMap *ConfigMap, pipeline *Pipeline) (*HttpServer, error) {
	return &HttpServer{
		pipeline:  pipeline,
		configMap: configMap,
	}, nil
}
