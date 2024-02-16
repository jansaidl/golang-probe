package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync/atomic"
	"syscall"
	"time"
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

			if r.URL.Query().Get("action") == "write" {
				d := r.URL.Query().Get("dir")

				f, err := os.Create(path.Join(d, fmt.Sprintf("file_%d", time.Now().Unix())))
				if err != nil {
					fmt.Fprintf(w, "err: %s\n", err.Error())

				} else {
					defer f.Close()
					m := md5.New()
					r := rand.New(rand.NewSource(time.Now().UnixMicro()))
					for i := 0; i < 1000; i++ {
						f.Write(m.Sum([]byte(fmt.Sprintf("aaaaa%d", r.Uint32()))))
					}
				}

				fmt.Fprintf(w, "\n------\n\n")
				printDir(w, d, "")

			}

			fmt.Fprintf(w, "\n------\n\n")
			printDir(w, "/mnt", "")
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

func printDir(w io.Writer, dirName string, prefix string) {
	dir, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(w, "%s: err: %s\n ", dirName, err.Error())
		return
	}
	for _, d := range dir {
		i, err := os.Stat(path.Join(dirName, d.Name()))
		if err != nil {
			fmt.Fprintf(w, "%s%s: %s\n ", prefix, d.Name(), err.Error())
			continue
		}
		if d.IsDir() {
			fmt.Fprintf(w, "%s%s dir\n ", prefix, d.Name())
			printDir(w, path.Join(dirName, d.Name()), prefix+" ")
		} else {
			fmt.Fprintf(w, "%s%s %d\n ", prefix, d.Name(), i.Size())
		}
	}
	return
}
