package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rakyll/hey/requester"
)

type worker struct {
	Address  string
	Index    string
	Buf      string
	Runs     int64
	Duration time.Duration
	Avg      time.Duration
	Cancel   context.CancelFunc
}

//go:embed list.html
var list string
var listTemplate *template.Template

func init() {
	listTemplate = template.Must(template.New("").Parse(list))
}

func (w *worker) stop() {
	w.Cancel()
}

func (w *worker) run(ctx context.Context) {
	cancelCtx, cancel := context.WithCancel(ctx)
	w.Cancel = cancel

	for {
		w.Runs++
		start := time.Now()
		result := bytes.NewBuffer(nil)
		hey(ctx, mustRequest(http.NewRequest("GET", w.Address, nil)), result)
		w.Buf = result.String()
		w.Duration += time.Since(start)
		if w.Runs > 0 {
			w.Avg = time.Duration(int64(w.Duration) / w.Runs)
		}
		select {
		case <-cancelCtx.Done():
			return
		default:
			//case <-time.After(time.Second * 5):
		}

	}
}

func mustRequest(r *http.Request, err error) *http.Request {
	if err != nil {
		panic(err)
	}
	return r
}

var workers []*worker

func main() {
	log.Println("starting")
	ctx, cancel := context.WithCancel(context.Background())

	rand.Seed(time.Now().Unix())

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		<-interrupt
		cancel()
	}()

	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := r.URL.Query().Get("do")
			switch action {
			case "addProjectInfo":
				addProjectInfo(ctx)
				return
			case "addApi":
				addApi(ctx)
				return
			case "remove":
				index := r.URL.Query().Get("index")
				for heyIndex, heyWorker := range workers {
					if heyWorker.Index == index {
						workers = append(workers[0:heyIndex], workers[heyIndex+1:]...)
						heyWorker.stop()
						fmt.Println("remove index: ", index)
						break
					}
				}
			}
			listTemplate.Execute(w, workers)
			return
		})); err != nil {
			log.Fatal(err)
		}
	}()

	for i := 0; i < 1; i++ {
		addProjectInfo(ctx)
		addProjectInfo(ctx)
		addProjectInfo(ctx)
		addProjectInfo(ctx)
		addProjectInfo(ctx)
		addApi(ctx)
		addApi(ctx)
		addApi(ctx)
		addApi(ctx)
		addApi(ctx)
	}

	<-ctx.Done()

}

func addProjectInfo(ctx context.Context) {
	chromWorker := newWorker("https://web-api.zerops.io/api/articles")
	workers = append(workers, chromWorker)
	go chromWorker.run(ctx)

}

func addApi(ctx context.Context) {
	chromWorker := newWorker("https://zerops.io/projectinfo")
	workers = append(workers, chromWorker)
	go chromWorker.run(ctx)
}

func newWorker(address string) *worker {
	return &worker{
		Address: address,
		Index:   strconv.Itoa(rand.Int()),
	}
}

func hey(ctx context.Context, req *http.Request, writer io.Writer) {
	w := &requester.Work{
		Request:            req,
		RequestBody:        nil,
		N:                  500,
		C:                  10,
		QPS:                0,
		Timeout:            20,
		DisableCompression: true,
		DisableKeepAlives:  true,
		DisableRedirects:   true,
		H2:                 false,
		ProxyAddr:          nil,
		Output:             "",
		Writer:             writer,
	}
	w.Init()
	go func() {
		<-ctx.Done()
		w.Stop()
	}()
	w.Run()

}
