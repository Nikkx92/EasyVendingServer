package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusOK, Response{Text: te})
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
