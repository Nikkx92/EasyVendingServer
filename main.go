package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

import (
	"context"
)

var (
	buf    bytes.Buffer
	logger = log.New(&buf, "server: ", log.Ltime+log.Lshortfile)
	jobs   = NewCancelMap()
	ca     = cache.New(1*time.Minute, 1*time.Minute)
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
	AutoMode     bool   `json:"AutoMode"`
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
	RespFns string   `json:"respFns"`
	DataKit []string `json:"text"`
	Buf     string   `json:"buffer"`
}

type CancelMap struct {
	sync.Mutex
	internal map[string]context.CancelFunc
}

func NewCancelMap() *CancelMap {
	return &CancelMap{
		internal: make(map[string]context.CancelFunc),
	}
}

func (c *CancelMap) Get(key string) (context.CancelFunc, bool) {
	c.Lock()
	defer c.Unlock()
	cancel, ok := c.internal[key]
	return cancel, ok
}

func (c *CancelMap) Set(key string, cancel context.CancelFunc) {
	c.Lock()
	defer c.Unlock()
	c.internal[key] = cancel
}

func (c *CancelMap) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.internal, key)
}

func backgroundWork(ctx context.Context, id string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Горутина %s остановлена", id)
			return
		case <-ticker.C:
			log.Printf("Горутина %s работает...", id)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	r := gin.Default()
	r.POST("/sendToKitVend", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//time.Sleep(2 * time.Second)

		d := toKitVending(req, c)
		c.JSON(http.StatusOK, Response{DataKit: d, Buf: buf.String()})
	})

	r.POST("/sendToFns", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		compl := toFns(req, c)
		c.JSON(http.StatusOK, Response{RespFns: compl, Buf: buf.String()})
	})

	r.POST("/start", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if _, exists := jobs.Get(req.Data.INN); exists {
			c.JSON(200, gin.H{"message": fmt.Sprintf("Горутина %s уже запущена", req.Data.INN)})
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		jobs.Set(req.Data.INN, cancel)

		// Запускаем горутину
		go backgroundWork(ctx, req.Data.INN)

		c.JSON(200, gin.H{"message": fmt.Sprintf("Горутина %s запущена", req.Data.INN)})
	})
	r.POST("/stop", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		cancel, exists := jobs.Get(req.Data.INN)
		if !exists {
			c.JSON(404, gin.H{"error": fmt.Sprintf("Горутина %s не найдена", req.Data.INN)})
			return
		}
		cancel()
		jobs.Delete(req.Data.INN)

		c.JSON(200, gin.H{"message": fmt.Sprintf("Горутина %s остановлена", req.Data.INN)})
	})

	if err := r.Run("192.168.1.88:8080"); err != nil {
		logger.Println("ошибка запуска сервера")
	}
}
