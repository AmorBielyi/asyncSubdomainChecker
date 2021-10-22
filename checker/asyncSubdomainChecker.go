package checker

import (
	"fmt"
	"github.com/amorbielyi/asyncSubdomainChecker/models"
	"net/http"
	"sync"
)

// Struct that represents Asynchronous Subdomain Checker
type asyncSubdomainChecker struct {

	// Let's keep stuff synchronized
	wg sync.WaitGroup

	// intermediate channel which holds result
	resultCh chan models.Result

	// slice of subdomains
	payload []string

	// slice of results
	results []models.Result
}

func NewAsyncSubdomainChecker(payload []string) *asyncSubdomainChecker {
	return &asyncSubdomainChecker{
		resultCh: make(chan models.Result),
		payload:  payload,
	}
}

// 1) Sends result as string into channel,
// 2) close channel,
// 3) receives result from channel.
// 4) returns slice of appended channel values.
func (checker *asyncSubdomainChecker) GetResults() []models.Result {
	checker.sendResults()
	checker.closeResults()
	checker.receiveResults()
	return checker.results
}

// Calls http.Get() on each subdomain URL via goroutine closure concurrently
// and sends string as result into channel.
// The result could be either:
//		Result{"subdomainName", "up", "httpStatusCode"} if subdomain is reachable
//      or
//      Result{"subdomainName", "down", 0}, if subdomain is not reachable

func (checker *asyncSubdomainChecker) sendResults() {
	for _, url := range checker.payload {
		checker.wg.Add(1)
		go func(s string) {
			defer checker.wg.Done()
			if resp, err := http.Get(fmt.Sprintf("https://%s", s)); err != nil {
				checker.resultCh <- models.Result{
					Subdomain: s,
					Status:    "down",
					Code:      0,
				}
			} else if resp, err = http.Get(fmt.Sprintf("http://%s", s)); err != nil {

				checker.resultCh <- models.Result{
					Subdomain: s,
					Status:    "down",
					Code:      0,
				}
			} else {
				checker.resultCh <- models.Result{
					Subdomain: s,
					Status:    "up",
					Code:      resp.StatusCode,
				}
			}
		}(url)
	}
}

// Iterates over channel in order to append and store received results
func (checker *asyncSubdomainChecker) receiveResults() {
	for msg := range checker.resultCh {
		checker.results = append(checker.results, msg)
	}
}

// Closes channel
func (checker *asyncSubdomainChecker) closeResults() {
	go func() {
		checker.wg.Wait()
		close(checker.resultCh)
	}()
}
