// Command race fetches the same file from several mirrors and prints whichever answers first.
//
//	go run ./examples/race 'https://go.dev/VERSION?m=text' 'https://golang.google.cn/VERSION?m=text'
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vardius/gollback/v2"
)

// fetchFirst asks every mirror at once. The first good answer wins and the other requests are cancelled.
func fetchFirst(ctx context.Context, client *http.Client, urls ...string) (string, error) {
	fetches := make([]gollback.AsyncFunc[string], len(urls))
	for i, url := range urls {
		fetches[i] = func(ctx context.Context) (string, error) {
			return get(ctx, client, url)
		}
	}

	return gollback.Race(ctx, fetches...)
}

func get(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%s: %w", url, err)
	}

	return string(body), nil
}

func main() {
	log.SetFlags(0)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := fetchFirst(ctx, http.DefaultClient, os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(body)
}
