package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Kitreq struct {
	CompanyId   string     `json:"CompanyId"`
	UserLogin   string     `json:"UserLogin"`
	PasswordKit string     `json:"PasswordKit"`
	INN         string     `json:"INN"`
	PasswordFns string     `json:"PasswordFns"`
	Date        string     `json:"Date"`
	Device      DeviceInfo `json:"Device"`
}

type Kitresp struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func main() {
	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		var req Kitreq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		co, te := mainRequest(req)
		c.JSON(http.StatusOK, Kitresp{Code: co, Text: te})
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
