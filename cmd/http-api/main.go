package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/maestre3d/dynamodb-tx-outbox/api"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/messaging"
)

const gracefulShutdownDuration = time.Second * 10

func main() {
	srv := api.NewHttpApi()
	api.InitEventApi()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	go func() {
		_ = messaging.DefaultEventBus.ListenAndServe()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	log.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownDuration)
	backgroundJobs := sync.WaitGroup{}
	backgroundJobs.Add(2)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	go func() {
		_ = srv.Shutdown(ctx)
		backgroundJobs.Done()
	}()
	go func() {
		_ = messaging.DefaultEventBus.Shutdown(ctx)
		backgroundJobs.Done()
	}()
	backgroundJobs.Wait()
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	os.Exit(0)
}
