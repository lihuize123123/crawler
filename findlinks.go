package main

import (
	"fmt"
	"log"
	"os"
)

type WorkTask struct {
	links []string
	deep  int
}

type LinkInfo struct {
	link string
	deep int
}

func crawl(url string) []string {
	fmt.Println(url)
	list, err := Extract(url)
	if err != nil {
		log.Print(err)
	}
	return list
}

func main() {
	worklist := make(chan WorkTask)    // lists of URLs, may have duplicates
	unseenLinks := make(chan LinkInfo) // de-duplicated URLs

	newWork := make(chan struct{})
	workDone := make(chan struct{})
	workCount := 0
	workDoneCount := 0
	linkCount := 0

	// determine whether all works were done
	go func() {
		for {
			select {
			case <-newWork:
				workCount++
			case <-workDone:
				workDoneCount++
				if workDoneCount == workCount {
					close(worklist)
					close(unseenLinks)
					return
				}
			}
		}
	}()

	// Add command-line arguments to worklist.
	go func() {
		newWork <- struct{}{}
		worklist <- WorkTask{os.Args[1:], 1}
	}()

	// Create 20 crawler goroutines to fetch each unseen link.
	for i := 0; i < 20; i++ {
		go func() {
			for link := range unseenLinks {
				foundLinks := crawl(link.link)
				go func() {
					worklist <- WorkTask{foundLinks, link.deep + 1}
				}()
			}
		}()
	}

	// The main goroutine de-duplicates worklist items
	// and sends the unseen ones to the crawlers.
	seen := make(map[string]bool)
	for list := range worklist {
		if list.deep > 2 {
			workDone <- struct{}{}
			continue
		}
		for _, link := range list.links {
			if !seen[link] {
				linkCount++
				seen[link] = true
				newWork <- struct{}{}
				unseenLinks <- LinkInfo{link, list.deep}
			}
		}
		workDone <- struct{}{}
	}

	fmt.Printf("Done, works count %v, links count %v\n", workCount, linkCount)
}
