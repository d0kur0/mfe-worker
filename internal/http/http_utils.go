package http

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

func getPagination(c echo.Context) (limit int, offset int) {
	requestLimit := c.QueryParam("limit")
	requestOffset := c.QueryParam("offset")

	limit, err := strconv.Atoi(requestLimit)
	if err != nil {
		limit = 20
	}

	offset, err = strconv.Atoi(requestOffset)
	if err != nil {
		offset = 0
	}

	return
}
