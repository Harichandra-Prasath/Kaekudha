package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Harichandra-Prasath/Kaekudha/internal/logger"
	"github.com/Harichandra-Prasath/Kaekudha/internal/pipeline"
)

type Config struct {
	serverAddr string
	name       string
	create     string
	join       string
}

func ParseConfig(args []string) (*Config, error) {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	cfg := &Config{}

	fs.StringVar(&cfg.serverAddr, "server", "", "Remote Server Address")
	fs.StringVar(&cfg.name, "name", "", "UserName")
	fs.StringVar(&cfg.create, "create", "", "Session Id to be Created")
	fs.StringVar(&cfg.join, "join", "", "Session Id to be Joined")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if cfg.name == "" || cfg.serverAddr == "" {
		return nil, fmt.Errorf("name and server cannot be empty")
	}

	if cfg.create == "" && cfg.join == "" {
		return nil, fmt.Errorf("either create or join is required")
	}

	if cfg.create != "" && cfg.join != "" {
		return nil, fmt.Errorf("create and join are mutually exclusive.")
	}

	if len(cfg.name) > 8 {
		return nil, fmt.Errorf("name cannot be more than 8 characters")
	}
	if len(cfg.join) > 8 {
		return nil, fmt.Errorf("join session id cannot be more than 8 characters")
	}
	if len(cfg.create) > 8 {
		return nil, fmt.Errorf("session create id cannot be more than 8 characters")
	}
	return cfg, nil
}

func main() {
	logger.InitialiseLogger()
	slog.Info("Logger Initialised")

	cfg, err := ParseConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}

	client, err := pipeline.GetNewClient(cfg.name, cfg.serverAddr)
	if err != nil {
		panic(fmt.Sprintf("error getting new client: %v", err))
	}

	err = client.Start(cfg.create, cfg.join)
	if err != nil {
		panic(fmt.Sprintf("error starting the client: %v", err))
	}
	slog.Info("Client Started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("End Signal Recieved", "Signal", sig.String())
	client.Stop()
}
