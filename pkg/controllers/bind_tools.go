package controllers

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, request any) error {
	buf := make([]byte, 5120)
	num, _ := c.Request.Body.Read(buf)
	reqBody := string(buf[0:num])
	c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(reqBody)))
	err := c.ShouldBindJSON(request)
	c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(reqBody)))
	return err
}

func BindJSONMap(c *gin.Context, request *map[string]any) error {
	buf := make([]byte, 5120)
	num, _ := c.Request.Body.Read(buf)
	reqBody := buf[0:num]
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	err := json.Unmarshal(reqBody, &request)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	return err
}

type MessageResponse struct {
	Message string `json:"message"`
}

func PaginationValues(limit int64, page int64, total int64) (numPages int64, nextCursor int64, prevCursor int64) {
	numPages = (total + limit - 1) / limit
	if page < numPages {
		nextCursor = page + 1
	}
	if page > 1 {
		prevCursor = page - 1
	}
	return
}
