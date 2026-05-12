package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/pelican-dev/panel/internal/api"
	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/config"
	"github.com/pelican-dev/panel/internal/database"
)

func main() {
	configPath := flag.String("config", "/etc/pelican/panel.yml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database error: %v\n", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis error: %v\n", err)
		os.Exit(1)
	}

	jwtKey := []byte("change-me-in-production")
	jwtManager := auth.NewJWTManager(jwtKey)

	api := api.NewAPI(db, cfg, jwtManager, rdb)

	go func() {
		fmt.Println("panel started")
		if err := api.Router.Run(":8080"); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("shutting down")
}
