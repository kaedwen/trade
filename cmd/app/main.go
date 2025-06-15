package main

import (
	"context"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/kaedwen/trade/pkg/app"
	"github.com/kaedwen/trade/pkg/app/utils"
)

func main() {
	ctx, end := context.WithCancel(context.Background())
	defer end()

	go utils.SigWatch(end, time.Second*5, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	app, err := app.NewApplication()
	if err != nil {
		log.Fatal("failed to create application", err)
	}

	app.Run(ctx)
}
