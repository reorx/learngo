package main

import (
	"fmt"
)

// Fetcher interface
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type result struct {
	url, body string
	urls      []string
	err       error
	depth     int
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	results := make(chan *result)
	fetched := make(map[string]bool)
	fetch := func(url string, depth int) {
		body, urls, err := fetcher.Fetch(url)
		results <- &result{url, body, urls, err, depth}
	}

	fetching := 0
	dispatch := func(url string, depth int) {
		fetching++
		go fetch(url, depth)
		fetched[url] = true
	}

	dispatch(url, depth)

	// 1 url is currently being fetched in background, loop while fetching
	for fetching > 0 {
		res := <-results
		fetching--

		// skip failed fetches
		if res.err != nil {
			fmt.Println(res.err)
			continue
		}

		fmt.Printf("found: %s %q\n", res.url, res.body)

		// follow links if depth has not been exhausted
		if res.depth <= 0 {
			continue
		}
		for _, u := range res.urls {
			// don't attempt to re-fetch known url, decrement depth
			if fetched[u] {
				continue
			}
			dispatch(u, res.depth-1)
		}
	}

	close(results)
}

func main() {
	Crawl("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f *fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := (*f)[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = &fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
