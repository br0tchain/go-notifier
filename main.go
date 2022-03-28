package main

import (
	"github.com/br0tchain/go-notifier/lib"
	"github.com/br0tchain/go-notifier/usecase"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	client lib.Client
)

func main() {
	begin := time.Now()
	client = lib.New(lib.SetVerbose(loadArgs()))
	notifier := usecase.NewNotifier(client)
	notifier.ReadInput()
	//handle quitting program
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("server started in %s", time.Since(begin))
	<-quit
	log.Printf("shutting down server ...")
	log.Printf("server exiting after %s", time.Since(begin))
}

func loadArgs() bool {
	if len(os.Args) > 1 {
		return os.Args[1] == "--verbose"
	}
	return false
}
