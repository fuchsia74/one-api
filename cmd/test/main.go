package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	glog "github.com/Laisky/go-utils/v5/log"
	"github.com/Laisky/zap"
	_ "github.com/joho/godotenv/autoload"
)

// main configures logging, listens for termination signals, and runs the regression harness.
func main() {
	logger, err := glog.NewConsoleWithName("oneapi-test", glog.LevelInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %+v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("test run failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("all tests passed")
}
