package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var (
	buf    bytes.Buffer
	logger = log.New(&buf, "server: ", log.Ltime+log.Lshortfile)
)

type DataApp struct {
	CompanyId    string `json:"CompanyId"`
	UserLogin    string `json:"UserLogin"`
	PasswordKit  string `json:"PasswordKit"`
	INN          string `json:"INN"`
	PasswordFns  string `json:"PasswordFns"`
	Date         string `json:"Date"`
	RefreshToken string `json:"RefreshToken"`
	Token        string `json:"Token"`
}

type MetaDetails struct {
	UserAgent string `json:"userAgent"`
}

type DeviceInfo struct {
	SourceDeviceID string      `json:"sourceDeviceId"`
	SourceType     string      `json:"sourceType"`
	AppVersion     string      `json:"appVersion"`
	MetaDetails    MetaDetails `json:"metaDetails"`
}

type Request struct {
	Data   DataApp    `json:"Data"`
	Device DeviceInfo `json:"Device"`
}

type Response struct {
	Text string `json:"text"`
	Buf  string `json:"buffer"`
}

func main() {
	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		te := mainRequest(req, c)
		c.JSON(http.StatusOK, Response{Text: te, Buf: buf.String()})
	})

	if err := r.Run(":8080"); err != nil {
		logger.Println("ошибка запуска сервера")
	}
}
