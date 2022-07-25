package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/chromedp/chromedp"
)

func main() {
	log.Println("starting")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		<-interrupt
		cancel()
	}()

	address, err := chrome(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Sprintf(address.String())
	log.Printf("Address: %s", address.String())

	var buf []byte

	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "image/png")
			w.Write(buf)

		})); err != nil {
			log.Fatal(err)
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(ctx context.Context, i int) {
			defer wg.Done()
			for {
				fmt.Println(i)
				allocatorCtx, _ := chromedp.NewRemoteAllocator(ctx, address.String())
				xCtx, _ := chromedp.NewContext(allocatorCtx)
				var title string
				actions := []chromedp.Action{
					chromedp.Navigate("https://zerops.io/"),
					chromedp.Title(&title),
					chromedp.WaitVisible("body > zw-app > div > zw-home-page > zw-section > div.__project-card-shift-wrap > div > div > div > zui-zerops-project-full-card > zui-wrap > zui-project-full-card > mat-card > div > div:nth-child(3) > zef-scroll > div.c-zef-scroll-area.__area > div > zui-wrap > zui-project-full-card-service-stacks > div:nth-child(7) > zui-service-stack-card > div > div.__ripple-wrap.ng-tns-c125-24 > div > div > zui-service-stack-basic-info > div > zui-basic-info-header > h3 > span > div > div:nth-child(1) > zef-fuse-highlight"),
				}
				if i == 9 {
					actions = append(actions, chromedp.FullScreenshot(&buf, 90))
				}

				if err := chromedp.Run(xCtx, actions...); err != nil {
					log.Println(err)
					return
				}
				//		log.Println(title)
				select {
				case <-ctx.Done():
					return
				default:
					//case <-time.After(time.Second * 5):

				}
			}
		}(ctx, i)
	}

	wg.Wait()

}

func chrome(ctx context.Context) (*url.URL, error) {
	cmd := exec.CommandContext(ctx, "/usr/bin/chromium-browser", "--no-sandbox", "--disable-gpu", "--headless", "--remote-debugging-port=9222")
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stderrPipe)

	scanned := make(chan bool)
	defer close(scanned)

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := cmd.Start(); err != nil {
			log.Printf("%v (%s)", err, "start command error")
			return
		}
		ok := <-scanned
		var err error
		if ok {
			err = cmd.Process.Release()
		} else {
			err = cmd.Process.Kill()
		}
		if err != nil {
			log.Println(err)
		}
	}()

	var addr *url.URL
	err = errors.New("address not found")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "DevTools listening on ") {
			line := strings.TrimPrefix(line, "DevTools listening on ")
			addr, err = url.Parse(line)
			if err != nil {
				err = fmt.Errorf("%v: invalid address: %s", err, line)
				log.Printf("%v (%s)", err, "start command error")
				return nil, err
			}
			err = nil
			break
		}
	}
	if err != nil || scanner.Err() != nil {
		scanned <- false
		<-done
		if scanner.Err() != nil {
			log.Println(err)
		}
		return nil, err
	}
	scanned <- true

	return addr, nil
}
