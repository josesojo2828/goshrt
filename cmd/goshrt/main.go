package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsojo/goshrt/internal/config"
	"github.com/jsojo/goshrt/internal/handler"
	"github.com/jsojo/goshrt/internal/service"
	"github.com/jsojo/goshrt/internal/store/postgres"
	"github.com/jsojo/goshrt/internal/store/redis"
	"github.com/jsojo/goshrt/internal/worker"
)

func main() {
	cfg := config.Load()

	pgStore, err := postgres.NewStore(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("connecting to postgres: %v", err)
	}
	defer pgStore.Close()

	redisCache, err := redis.NewCache(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("connecting to redis: %v", err)
	}
	defer redisCache.Close()

	svc := service.New(pgStore, redisCache)

	ctx := context.Background()
	count, err := svc.SeedCache(ctx)
	if err != nil {
		log.Printf("warning: cache seed: %v", err)
	} else {
		log.Printf("seeded %d URLs to cache", count)
	}

	clickWorker := worker.New(svc, cfg.ClickSyncInterval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go clickWorker.Start(ctx)

	h := handler.New(svc)
	addr := cfg.Port
	if addr[0] != ':' {
		addr = ":" + addr
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: h.Routes(),
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		clickWorker.Stop()
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("goshrt listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
