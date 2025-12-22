package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"net"
	"net/http"
	kitpb "test/api"
	"test/internal/fns"
	"test/internal/kitvending"
	"test/internal/services"
	"test/internal/store"
	grpcTransport "test/internal/transport/grpc"
	httpTransport "test/internal/transport/http"
	"time"
)

type Config struct {
	DSN      string
	GRPCAddr string
	HTTPAddr string
}

type App struct {
	cfg        Config
	pool       *pgxpool.Pool
	grpcServer *grpc.Server
	httpServer *http.Server
	lis        net.Listener
}

func New(cfg Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Start(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, a.cfg.DSN)
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return err
	}
	a.pool = pool

	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient := &http.Client{Transport: tr}

	fnsClient := fns.NewClient(httpClient, "https://lknpd.nalog.ru")
	kitClient := kitvending.NewClient(httpClient, "https://api2.kit-invest.ru")

	st := store.New(a.pool)
	authSvc := services.NewAuthService(st, fnsClient, kitClient)
	kitSvc := services.NewKitService(st, kitClient)

	httpHandler := httpTransport.NewHTTPHandler(authSvc)
	grpcHandler := grpcTransport.NewGRPCHandler(authSvc, kitSvc)

	lis, err := net.Listen("tcp", a.cfg.GRPCAddr)
	if err != nil {
		a.pool.Close()
		return err
	}
	a.lis = lis
	a.grpcServer = grpc.NewServer()
	kitpb.RegisterKitNalogServiceServer(a.grpcServer, grpcHandler)

	go func() {
		if err := a.grpcServer.Serve(lis); err != nil {
			a.pool.Close()
		}
	}()

	r := gin.Default()
	httpTransport.RegisterRouters(r, httpHandler)
	a.httpServer = &http.Server{
		Addr:    a.cfg.HTTPAddr,
		Handler: r,
	}
	go func() {
		_ = a.httpServer.ListenAndServe()
	}()
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.httpServer != nil {
		_ = a.httpServer.Shutdown(ctx)
	}

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	if a.pool != nil {
		a.pool.Close()
	}

	if a.lis != nil {
		_ = a.lis.Close()
	}
	return nil
}
