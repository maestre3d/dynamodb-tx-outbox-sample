package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/maestre3d/dynamodb-tx-outbox/api"
)

const gracefulShutdownDuration = time.Second * 10

func main() {
	srv := api.NewHttpApi()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	log.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownDuration)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	os.Exit(0)
}
