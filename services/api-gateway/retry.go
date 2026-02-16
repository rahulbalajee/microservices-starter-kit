package main

import (
	"log"
	"sync/atomic"
	"time"
)

func connectWithBackoff[T any](maxBackoff time.Duration, name string, ptr *atomic.Pointer[T], newClient func() (*T, error)) {
	backoff := time.Second
	for {
		client, err := newClient()
		if err != nil {
			log.Printf("%s client: %v (retry in %v)\n", name, err, backoff)
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}
		ptr.Store(client)
		log.Printf("%s service client connected\n", name)
		return
	}
}
