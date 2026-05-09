package main

import (
	"context"
	"fmt"
	"nexo/internal/hashring"
	"nexo/internal/network"
	"nexo/internal/store"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("NEXO running ✨")
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	defer stop()

	workers := []int{9091, 9092, 9093, 9094}

	for _, port := range workers {
		wg.Add(1)
		ss := store.New[string](100)

		go func() {
			s := network.Server{St: ss, Port: port}
			s.Start(ctx, &wg)
		}()
	}

	hr := hashring.New()
	for _, port2 := range workers {
		hr.AddNode(fmt.Sprintf("localhost:%d", port2), 5)
	}

	c := network.Coordinator{Port: 9090, Ring: hr}
	wg.Add(1)
	c.Start(ctx, &wg)

	wg.Wait()

	fmt.Println("Awaiting signal")
	<-ctx.Done()

	fmt.Println()
	fmt.Println(context.Cause(ctx))
	fmt.Println("exiting")
}
