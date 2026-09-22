package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dynastic/internal/config"
	"dynastic/internal/database"
	"dynastic/internal/handler"
	"dynastic/internal/middleware"
	"dynastic/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()
	log.Printf("已连接数据库 %s@%s:%d/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	repo := repository.New(db)
	h := handler.New(repo)
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           h.Router(middleware.CORS(cfg.CORSOrigins)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 在独立 goroutine 中启动服务，主协程负责优雅关闭。
	go func() {
		log.Printf("帝王世系图谱后端已启动，监听 http://localhost%s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务关闭出错: %v", err)
	}
	log.Println("服务已停止")
}
