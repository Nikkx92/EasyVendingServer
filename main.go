package main

import (
	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
	"log"
	"net/http"
)

type Kitreq struct {
	CompanyId string `json:"CompanyId"`
	UserLogin string `json:"UserLogin"`
	Password  string `json:"Password"`
	Date      string `json:"Date"`
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
		co, te := mainRequest(req.CompanyId, req.UserLogin, req.Password, req.Date)
		c.JSON(http.StatusOK, Kitresp{Code: co, Text: te})
	})

	addr := ":4433"
	log.Printf("serving HTTP/3 on https://localhost%v", addr)

	if err := http3.ListenAndServeTLS(addr, "cert.pem", "key.pem", r); err != nil {
		log.Fatal(err)
	}
}
