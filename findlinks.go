package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	worklist := make(chan WorkTask)    // lists of URLs, may have duplicates
	unseenLinks := make(chan LinkInfo) // de-duplicated URLs

	workDone := make(chan struct{})
	workCount := 0
	workDoneCount := 0
	linkCount := 0

	// determine whether all works were done
	go func() {
		for {
			select {
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
	workCount++
	go func() {
		worklist <- WorkTask{flag.Args(), 1}
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
				workCount++
				linkCount++
				seen[link] = true
				unseenLinks <- LinkInfo{link, list.deep}
			}
		}
		workDone <- struct{}{}
	}

	fmt.Printf("Done, works count %v, links count %v\n", workCount, linkCount)
}
