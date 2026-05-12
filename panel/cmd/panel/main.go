package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pelican-dev/panel/internal/api"
	"github.com/pelican-dev/panel/internal/auth"
	"github.com/pelican-dev/panel/internal/config"
	"github.com/pelican-dev/panel/internal/database"
	"github.com/pelican-dev/panel/internal/jobs"
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

	jwtKey := []byte(cfg.App.Secret)
	if len(jwtKey) == 0 {
		jwtKey = []byte(os.Getenv("PANEL_APP_SECRET"))
	}
	if len(jwtKey) == 0 {
		jwtKey = []byte("change-me-in-production")
	}
	jwtManager := auth.NewJWTManager(jwtKey)

	router := api.NewAPI(db, cfg, jwtManager, rdb)

	queue := jobs.NewQueue(rdb, db)
	go func() {
		if err := queue.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "queue error: %v\n", err)
		}
	}()

	srv := &http.Server{Addr: ":8080", Handler: router.Router}
	go func() {
		fmt.Println("panel started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	if rdb != nil {
		rdb.Close()
	}
}
