package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

func main() {

	var i uint64

	fmt.Println("starting")
	go func() {
		http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&i, 1)
			w.Header().Add("Content-type", "text/plain")
			fmt.Fprintf(w, "Hello %d ;o)\n\n", i)
			for _, e := range os.Environ() {
				fmt.Fprintf(w, "%s\n", e)
			}
		}))
	}()

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
