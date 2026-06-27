package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"helloGithub/src/internal/app"
	"helloGithub/src/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// 创建服务管理器（统一管理所有组件，为GUI预留接口）
	manager, err := app.NewServiceManager(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init service manager: %v\n", err)
		os.Exit(1)
	}

	log := manager.GetLogger()
	log.Info("GitHub Fast DNS starting...")

	// 启动所有服务
	if err := manager.Start(); err != nil {
		log.Fatal("failed to start services", "error", err)
	}

	log.Info("GitHub Fast DNS started successfully",
		"dns_addr", fmt.Sprintf("%s:%d", cfg.DNS.ListenAddr, cfg.DNS.ListenPort),
		"scan_interval", cfg.Scanner.Interval,
	)

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Info("shutting down...")

	// 停止所有服务
	if err := manager.Stop(); err != nil {
		log.Error("stop services", "error", err)
	}

	log.Info("GitHub Fast DNS stopped")
}
