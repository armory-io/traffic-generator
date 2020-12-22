package main

import (
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func fulfillRequest(url string) int {
	log.Info("Making requests to:" + url)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	if err != nil {
		log.Errorf("Finished with error: %v", err)
		return 0
	} else if resp.StatusCode > 400 {
		log.Errorf("Finished with status code over 400, code was: %d", resp.StatusCode)
		return 0
	}
	return 1
}
func addFixedRequests(reqCh chan int, maxRequests int) {
	log.Infof("Adding %d requests", maxRequests)
	completedRequests := 0
	for completedRequests < maxRequests {
		reqCh <- 1
		completedRequests++
	}
}

func addInifiniteRequests(reqCh chan int, maxRequests int, d time.Duration) {
	ticker := time.NewTicker(d)
	completedRequests := 0
	for range ticker.C {
		reqCh <- 1
		completedRequests++
		if maxRequests > 0 && completedRequests > maxRequests {
			ticker.Stop()
		}
	}
}

var (
	url                      string
	maxRequests, concurrency int
	requestInterval          time.Duration
)

func main() {
	flag.IntVar(&concurrency, "concurrency", 1, "how many requests to concurrently run")
	flag.StringVar(&url, "url", "http://localhost:8000", "host name")
	flag.IntVar(&maxRequests, "max-requests", 0, "Total number of requests")
	flag.DurationVar(&requestInterval, "request-interval", 0, "Time in millisconds between requests added to the channel")
	flag.Parse()

	log.Info("Traffic Generator started...")
	reqCh := make(chan int, concurrency)

	if maxRequests > 0 && requestInterval == 0 {
		go addFixedRequests(reqCh, maxRequests)
	} else if requestInterval > 0 {
		go addInifiniteRequests(reqCh, maxRequests, requestInterval)
	} else {
		go addInifiniteRequests(reqCh, maxRequests, time.Millisecond*10)
	}

	log.Info("Looping through requests")
	for range reqCh {
		fulfillRequest(url)
	}
}
