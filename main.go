package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

var wg sync.WaitGroup

func downloadFile(ctx context.Context, url string, ch chan string) {

	defer wg.Done()

	// creates an HTTP GET request for the url, and ataches context --> if the context expires/timeout, this request will auto cancel
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		ch <- fmt.Sprintf("ERROR: %v", err)
		return
	}

	// sends request and waits for response
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		ch <- fmt.Sprintf("ERROR downloading %s: %v", url, err)
		return
	}

	defer resp.Body.Close() // closing response body

	fileName := fmt.Sprintf("file-%d.html", time.Now().UnixNano()) // UnixNano ensures a unique filename in nanoseconds.
	outFile, err := os.Create(fileName)

	if err != nil {
		ch <- fmt.Sprintf("ERROR creating file: %v", err)
		return
	}

	defer outFile.Close()

	//
	total := resp.ContentLength
	bar := pb.Full.Start64(total)

	reader := bar.NewProxyReader(resp.Body)
	//

	_, err = io.Copy(outFile, reader) // copy responce body to your file

	if err != nil {
		ch <- fmt.Sprintf("ERROR writing to file: %v", err)
		return
	}

	bar.Finish()

	ch <- fmt.Sprintf("Downloded %s -> %s", url, fileName) // send msg
}

func main() {

	urls := []string{
		"https://golang.org",
		"https://example.com",
		"https://httpbin.org/delay/2", // slow URL to test concurrency
	}

	ch := make(chan string) // create channel ch

	// creats a context with 10 sec timeout --> if download takes too long, it'll be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start downloads
	for _, url := range urls {
		wg.Add(1)
		go downloadFile(ctx, url, ch)
	}

	// closing channel when done by waiting in another go routine
	go func() {
		wg.Wait()
		close(ch)
	}()

	// print downloaded
	for msg := range ch {
		fmt.Println(msg)
	}

	fmt.Println("All downloads complete")
}
