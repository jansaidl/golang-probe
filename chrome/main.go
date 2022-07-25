package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	allocatorCtx, _ := chromedp.NewRemoteAllocator(ctx, address.String())
	runCtx, _ := chromedp.NewContext(allocatorCtx)

	for {
		var title string
		actions := []chromedp.Action{
			chromedp.Navigate("https://www.google.com/"),
			chromedp.Title(&title),
		}
		if err := chromedp.Run(runCtx, actions...); err != nil {
			log.Println(err)
			return
		}
		log.Println(title)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):

		}
	}
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
