// Command retry fetches a URL, retrying with a growing pause between attempts.
//
//	go run ./examples/retry https://example.com
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

// getWithBackoff makes up to attempts requests. Before each retry it waits, doubling the pause every time,
// and stops waiting as soon as ctx is done.
func getWithBackoff(ctx context.Context, client *http.Client, url string, attempts int, pause time.Duration) (string, error) {
	var last error // the previous request's
	return gollback.Retry(ctx, attempts, func(ctx context.Context) (string, error) {
		if last != nil {
			select {
			case <-time.After(pause):
				pause *= 2
			case <-ctx.Done():
				// Retry joins ctx's error with this one, so the reason for retrying isn't lost.
				return "", last
			}
		}

		body, err := get(ctx, client, url)
		last = err
		return body, err
	})
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
	if len(os.Args) != 2 {
		log.Fatal("usage: retry <url>")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body, err := getWithBackoff(ctx, http.DefaultClient, os.Args[1], 5, 200*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(body)
}
