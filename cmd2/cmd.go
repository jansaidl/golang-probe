package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main () {
	ctx, cancel := context.WithCancel(context.Background())


	for _, v := range os.Environ() {
		fmt.Println(v)
	}

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
