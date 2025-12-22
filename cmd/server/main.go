package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"test/internal/app"
	"time"
)

/*
	func main() {
		runtime.GOMAXPROCS(1)
		services.SlogLogger()

		pool, err := pgxpool.New(context.Background(),
			"postgres://postgres:sonne@192.168.1.46:5432/postgres?sslmode=disable",
		)
		if err != nil {
			fmt.Println(err)
		}
		defer pool.Close()

		lis, err := net.Listen("tcp", "192.168.1.88:50051")
		if err != nil {
			fmt.Println(err)
		}
		grpcServer := grpc.NewServer()
		kitpb.RegisterKitNalogServiceServer(grpcServer, grpc2.NewServer(pool))
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				fmt.Println(err)
			}
		}()

		r := gin.Default()
		//h := NewHTTPHandler(pool)

		r.POST("/test", func(c *gin.Context) {
			time.Sleep(4 * time.Second)
		})

		r.POST("/sendToKitVend", services.AuthMiddleware(), func(c *gin.Context) {
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

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
			data := services.InitData(srvCtx, req, userId)
			d, message := services.ToKitVending(srvCtx, data)
			if message != "" {
				c.JSON(200, services.ResponseKit{Data: nil, Message: message})
			} else {
				c.JSON(http.StatusOK, services.ResponseKit{Data: d})
			}
		})
		r.POST("/sendToFns", services.AuthMiddleware(), func(c *gin.Context) {
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
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
			data := services.InitData(srvCtx, req, userId)
			services.AddDrinksDb(srvCtx, data, req.DataKit)
			message, modifiedCustomer := services.Check(srvCtx, data, req.DataKit)
			services.FnsTokenUpdate(srvCtx, modifiedCustomer, userId)
			c.JSON(http.StatusOK, services.ResponseFns{Message: message})
		})

		r.GET("/start", services.AuthMiddleware(), func(c *gin.Context) {
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var req services.SingleRequest
			v, ok := c.Get("userId")
			if !ok {
				fmt.Println("userId is not int64")
			}
			userId, ok := v.(int64)
			data := services.InitData(srvCtx, req, userId)
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
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var req services.SingleRequest
			v, ok := c.Get("userId")
			if !ok {
				fmt.Println("userId is not int64")
			}
			userId, ok := v.(int64)
			data := services.InitData(srvCtx, req, userId)

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

		r.GET("/sales", services.AuthMiddleware(), func(c *gin.Context) {
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var req services.SingleRequest

			v, ok := c.Get("userId")
			if !ok {
				fmt.Println("userId is not int64")
			}
			userId, ok := v.(int64)
			data := services.InitData(srvCtx, req, userId)
			history := services.GetSales(srvCtx, data)
			c.JSON(200, gin.H{"data": history})
		})

		r.POST("/auth", func(c *gin.Context) {
			srvCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var req services.Request
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			jwtToken, err := services.Authorization(srvCtx, req)
			if err != nil {
				c.JSON(200, services.AuthResponse{IsValid: false, Message: err.Error()})
			}

			if jwtToken != "" {
				c.JSON(200, services.AuthResponse{IsValid: true, Message: jwtToken})
			} else {
				err := services.ValidLoginKit(req)
				if err != nil {
					c.JSON(200, services.AuthResponse{IsValid: false, Message: err.Error()})
				}
				tokens, err := services.GetRefreshToken(req)
				if err != nil {
					c.JSON(200, services.AuthResponse{IsValid: false, Message: err.Error()})
				}
				jwtToken = services.AddCustomer(req, tokens.RefreshToken, tokens.Token)
				c.JSON(200, services.AuthResponse{IsValid: true, Message: jwtToken})

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
*/
func main() {
	cfg := app.Config{
		DSN:      "postgres://postgres:sonne@192.168.1.46:5432/postgres?sslmode=disable",
		GRPCAddr: "192.168.1.88:50051",
		HTTPAddr: ":8080",
	}

	a := app.New(cfg)

	ctx := context.Background()
	if err := a.Start(ctx); err != nil {
		panic(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = a.Stop(stopCtx)
}
