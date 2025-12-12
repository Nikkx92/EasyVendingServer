package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"test/internal/services"
)

import (
	"context"
)

func main() {
	runtime.GOMAXPROCS(1)
	r := gin.Default()

	r.POST("/sendToKitVend", services.AuthMiddleware(), func(c *gin.Context) {
		var req services.SingleRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		data := services.InitData(req, userId)
		d := services.ToKitVending(data)
		c.JSON(http.StatusOK, services.ResponseKit{Data: d, Buf: services.Buf.String()})
	})
	r.POST("/sendToFns", services.AuthMiddleware(), func(c *gin.Context) {
		var req services.SingleRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		data := services.InitData(req, userId)
		services.AddDrinksDb(data, req.DataKit)
		message, modifiedCustomer := services.Check(data, req.DataKit)
		services.FnsTokenUpdate(modifiedCustomer, userId)
		c.JSON(http.StatusOK, services.ResponseFns{Message: message, Buf: services.Buf.String()})
	})

	r.GET("/start", services.AuthMiddleware(), func(c *gin.Context) {
		var req services.SingleRequest
		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		data := services.InitData(req, userId)
		if _, exists := services.Jobs.Get(data.INN); exists {
			c.JSON(200, gin.H{"message": fmt.Sprintf("Автоматическая отправка для %s уже включена", data.INN)})
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		services.Jobs.Set(data.INN, cancel)

		// Запускаем горутину
		go services.BackgroundWork(ctx, data, userId)
		services.WorkingAuto[data.INN]++
		c.JSON(200, gin.H{"message": fmt.Sprintf("Автоматическая отправка для %s включена успешно", data.INN)})
	})
	r.GET("/stop", services.AuthMiddleware(), func(c *gin.Context) {
		var req services.SingleRequest
		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		data := services.InitData(req, userId)

		cancel, exists := services.Jobs.Get(data.INN)
		if !exists {
			c.JSON(404, gin.H{"error": fmt.Sprintf("ИНН %s не найден", data.INN)})
			return
		}
		cancel()
		services.Jobs.Delete(data.INN)
		delete(services.WorkingAuto, data.INN)
		c.JSON(200, gin.H{"message": fmt.Sprintf("Автоматическая отправка для %s остановлена", data.INN)})
	})

	r.GET("/history", services.AuthMiddleware(), func(c *gin.Context) {
		var req services.SingleRequest

		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		data := services.InitData(req, userId)
		history := services.GetHistory(data)
		c.JSON(200, gin.H{
			"data": history,
		})
	})

	r.POST("/auth", func(c *gin.Context) {
		var req services.Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		exist, jwtToken := services.Authorization(req)
		if exist {
			c.JSON(200, services.AuthResponse{IsValid: true, Message: jwtToken})
		} else {
			message := services.ValidLoginKit(req)
			if message == "" {
				//rt, t := getRefreshToken(req.Device, req.Data.INN, req.Data.PasswordFns)
				rt := ""
				t := "tttttttt"
				if t == "" {
					c.JSON(200, services.AuthResponse{IsValid: false, Message: rt})
				} else {
					jwtToken = services.AddCustomer(req, rt, t)
					c.JSON(200, services.AuthResponse{IsValid: true, Message: jwtToken})
				}
			} else {
				c.JSON(200, services.AuthResponse{IsValid: false, Message: message})
			}
		}
	})

	r.GET("/countOfAutoMode", services.AuthMiddleware(), func(c *gin.Context) {
		v, ok := c.Get("userId")
		if !ok {
			fmt.Println("userId is not int64")
		}
		userId, ok := v.(int64)
		if userId != 35 {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied: invalid user ID"})
			return
		}
		c.JSON(200, gin.H{"data": services.WorkingAuto})
	})

	if err := r.Run("192.168.1.88:8080"); err != nil {
		fmt.Println("error start server")
	}
}
