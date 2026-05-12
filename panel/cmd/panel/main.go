package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	jwtKey := []byte("change-me-in-production")
	jwtManager := auth.NewJWTManager(jwtKey)

	api := api.NewAPI(db, cfg, jwtManager)

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
