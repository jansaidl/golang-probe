package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync/atomic"
	"syscall"
)

func main() {

	var i, j uint64

	fmt.Println("starting")
	go func() {
		http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&i, 1)
			w.Header().Add("Content-type", "text/plain")
			fmt.Fprintf(w, "Hello i %d / %d;o)\n\n", i, j)
			for _, e := range os.Environ() {
				fmt.Fprintf(w, "%s\n", e)
			}
		}))
	}()
	go func() {
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&j, 1)
			w.Header().Add("Content-type", "text/plain")
			fmt.Fprintf(w, "Hello j %d / %d;o)\n\n", i, j)

			fmt.Fprintf(w, "\n------\n\n")
			printDir(w, "/tmp/sharedstorage0")
			fmt.Fprintf(w, "\n------\n\n")

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

func printDir(w io.Writer, dirName string) {
	dir, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(w, "%s: err: %s\n ", dirName, err.Error())
		return
	}
	for _, d := range dir {
		i, err := os.Stat(path.Join(dirName, d.Name()))
		if err != nil {
			fmt.Fprintf(w, "%s: %s\n ", d.Name(), err.Error())
			continue
		}
		if d.IsDir() {
			fmt.Fprintf(w, "%s dir\n ", d.Name())
		} else {
			fmt.Fprintf(w, "%s %d\n ", d.Name(), i.Size())
		}
	}
	return
}
