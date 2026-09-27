package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type LocationPayload struct {
	DriverID string  `json:"driver_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type MatchPayload struct {
	RiderLat float64 `json:"rider_lat"`
	RiderLng float64 `json:"rider_lng"`
}

func main() {
	concurrency := flag.Int("concurrency", 50, "Number of concurrent worker goroutines")
	duration := flag.Duration("duration", 10*time.Second, "Test duration")
	serverURL := flag.String("url", "http://localhost:8080", "Server endpoint")
	flag.Parse()

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
		},
	}

	var totalRequests uint64
	var successRequests uint64
	var failedRequests uint64

	fmt.Printf("Starting stress test against %s\n", *serverURL)
	fmt.Printf("Concurrency: %d workers | Duration: %s\n\n", *concurrency, *duration)

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	startTime := time.Now()

	for workerID := 0; workerID < *concurrency; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			for {
				select {
				case <-stopChan:
					return
				default:
					// 70% location updates (writes), 30% rider match requests (reads)
					isWrite := rng.Float64() < 0.7
					var err error

					if isWrite {
						payload := LocationPayload{
							DriverID: fmt.Sprintf("driver-%d", rng.Intn(2000)),
							Lat:      43.60 + rng.Float64()*0.20,
							Lng:      -79.50 + rng.Float64()*0.30,
						}
						body, _ := json.Marshal(payload)
						resp, reqErr := client.Post(*serverURL+"/drivers/location", "application/json", bytes.NewReader(body))
						if reqErr == nil {
							resp.Body.Close()
							if resp.StatusCode == http.StatusAccepted {
								atomic.AddUint64(&successRequests, 1)
							} else {
								atomic.AddUint64(&failedRequests, 1)
							}
						} else {
							err = reqErr
						}
					} else {
						payload := MatchPayload{
							RiderLat: 43.60 + rng.Float64()*0.20,
							RiderLng: -79.50 + rng.Float64()*0.30,
						}
						body, _ := json.Marshal(payload)
						resp, reqErr := client.Post(*serverURL+"/match", "application/json", bytes.NewReader(body))
						if reqErr == nil {
							resp.Body.Close()
							if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
								atomic.AddUint64(&successRequests, 1)
							} else {
								atomic.AddUint64(&failedRequests, 1)
							}
						} else {
							err = reqErr
						}
					}

					atomic.AddUint64(&totalRequests, 1)
					if err != nil {
						atomic.AddUint64(&failedRequests, 1)
					}
				}
			}
		}(workerID)
	}

	time.Sleep(*duration)
	close(stopChan)
	wg.Wait()

	totalTime := time.Since(startTime).Seconds()
	total := atomic.LoadUint64(&totalRequests)
	success := atomic.LoadUint64(&successRequests)
	failed := atomic.LoadUint64(&failedRequests)
	rps := float64(total) / totalTime

	fmt.Println("--- Load Test Results ---")
	fmt.Printf("Total Requests:   %d\n", total)
	fmt.Printf("Successful:       %d\n", success)
	fmt.Printf("Failed:           %d\n", failed)
	fmt.Printf("Throughput (RPS): %.2f req/sec\n", rps)
}
