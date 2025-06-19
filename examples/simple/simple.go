// Copyright © 2022 Atonal Authors
//

package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/hedzr/is/term/color"
	"github.com/hedzr/progressbar/v2"
)

func main() {
	color.Hide()
	defer color.Show()

	var wg sync.WaitGroup
	wg.Add(1)

	// _, _ = fmt.Fprintln(progressbar.New(), "Starting....")

	req, _ := http.NewRequest("GET", "https://dl.google.com/go/go1.14.2.src.tar.gz", nil) //nolint:gocritic
	resp, _ := http.DefaultClient.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}(resp.Body) //nolint:govet //just a demo

	f, _ := os.OpenFile("go1.14.2.src.tar.gz", os.O_CREATE|os.O_WRONLY, 0o644)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}(f)

	const BUFFERSIZE = 4096
	buf := make([]byte, BUFFERSIZE)

	// s.w = io.MultiWriter(f, bar)

	progressbar.Add(
		resp.ContentLength,
		"downloading go1.14.2.src.tar.gz", // fmt.Sprintf("downloading %v", s.fn),
		// progressbar.WithBarSpinner(14),
		// progressbar.WithBarStepper(3),
		progressbar.WithBarStepper(0),
		progressbar.WithBarOnCompleted(func(bar progressbar.MiniResizeableBar) {
			wg.Done()
		}),
		progressbar.WithBarWorker(func(bar progressbar.MiniResizeableBar, exitCh <-chan struct{}) (stop bool) {
			for {
				n, err := resp.Body.Read(buf)
				if err != nil && !errors.Is(err, io.EOF) {
					return
				}
				if n == 0 {
					break
				}

				select {
				case <-exitCh:
					return
				default: // avoid block at <-exitCh
				}

				if _, err = io.MultiWriter(f, bar).Write(buf[:n]); err != nil {
					return
				}
			}

			// _, _ = io.Copy(io.MultiWriter(f, bar), resp.Body)
			return
		}),
	)

	wg.Wait()
}
