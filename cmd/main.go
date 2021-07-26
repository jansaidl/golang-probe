package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main()  {
	fmt.Println("starting")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		<-interrupt
		cancel()
	}()
	fmt.Println("running")
	<-ctx.Done()
	fmt.Println("finished")
}



