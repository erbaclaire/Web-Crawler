package main

import (
	"src/html-wrangler"
	"src/book"
	"src/linkedlist"
	"src/stealing-queue"
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	"runtime"
	"sync/atomic"
	"math/rand"
)

// printUsage prints the required and optional command line arguments that should / can be specified for running the program.
func printUsage() {
	fmt.Println("Usage: scraper.go [-p=[number of threads]]\n\t-p=[number of threads] = An optional flag to run the scraper in its parallel version.")
}

// getBookUrls is used by both the sequential and parallel versions of the program to scrape
// all of the individual book webpages for a given genre and returns those urls in a slice
// so that those pages can be scraped.
func getBookUrls(endOfUrl string) []string {
	baseUrl := "http://books.toscrape.com/catalogue/category/books/"
	url := baseUrl + endOfUrl
	urlResponse := html.GetBodyResponse(url)
	bookUrls := html.GetBookUrls(urlResponse, url)	
	return bookUrls
}

// scrapeBookUrls scrapes urls for the book's details and determines if that book is the current
// lowest priced book for the genre in the linked-list dictionary.
// If it is, it updates the dictionary entry.
func scrapeBookUrl(bookUrl string, genre string, bookDict linkedlist.List) {
	bookUrlResponse := html.GetBodyResponse(bookUrl)
	book := html.GetBookDetails(bookUrlResponse)
	bookDict.Add(genre, book)
}

// bookDetailWorker is a goroutine that reads from the bookUrls channel.
// bookDetailWorker grabs a book url and scrapes that web page for book details.
// bookDetailWorker adds the genre's book to the linked list dictionary if there are no books for that given genre.
// if the genre is already in the linked list dictionary and the worker's book has a lower price, bookDetailWorker will 
// update the linked list dictionary's genre's entry to be the new book.
// bookDetailWorker will exit when there are no more book urls to consume.
func bookDetailWorker(threads int, bookDict linkedlist.List, genre string, bookUrls <-chan string, doneBookDetailWorker chan<- bool, qs []queue.Queue, i int, availableToSteal []bool, nothingLeftToStealCount *int64) {
	
	bookBase := "http://books.toscrape.com/catalogue/"

	for true {
		bookUrl, more := <- bookUrls // Grab from bookUrl channel.
		if more { // If get a url, add it to your stealing queue.
			bookUrl = bookBase + bookUrl[9:]
			qs[i].PushBottom(bookUrl)			
		} else {
			break // When no more book urls, break from this infinite for loop.
		}
	}

	// Pop book urls from your stealing queue if there are more.
	// If not, steal urls from another goroutine's stealing queue until that queue is also empty.
	// Then exit.
	for true {
		url := qs[i].PopBottom()
		if url == "nil" {
			// Alert that you don't have anything to steal and update the nothing left to steal count.
			availableToSteal[i] = false
			atomic.AddInt64(nothingLeftToStealCount, int64(1))
			// Steal from other threads until no queue has anything left to steal
			for atomic.LoadInt64(nothingLeftToStealCount) < int64(threads) {
				rand.Seed(int64(threads))
				randomThread := rand.Intn(threads)
				if availableToSteal[randomThread] {
					url := qs[randomThread].PopTop()
					if url == "nil" { // There is always the possibility that the queue being stolen from doesn't have anything anymore.
						continue // In this case, try to steal from another queue.
					} else { // Otherwise, scrape the stolen url.
						// fmt.Println(url, randomThread, i)
						scrapeBookUrl(url, genre, bookDict)
					}
				}
			}
			break			
		} else {
			scrapeBookUrl(url, genre, bookDict)
		}
	}
	
	// Alert that the book detail worker is done.
	doneBookDetailWorker <- true
	return
}

// worker is a goroutine that reads from the genreUrls channel.
// worker grabs a genre url from the genreUrls channel and scrapes the url to find all links to individual book urls for further scraping.
// worker then pushes the individual book urls to a book details channel.
// worker will continue to grab genre urls until there are no more genre urls to consume.
// worker will wait for all of its book detail workers to exit before it exits.
func worker(threads int, doneWorker chan<- bool, genreUrls <-chan string, bookDict linkedlist.List) {

	var detailWorkerCounter int
	doneBookDetailWorker := make(chan bool)
	defer close(doneBookDetailWorker)

	// Initial placing book urls in to a channel for bookDetailWorkers to consume.
	bookUrlStream := func(done <- chan interface{}, bookUrlSlice []string) chan string {
		bookUrls := make(chan string)
		go func() {
			defer close(bookUrls)
			for _, bookUrl := range bookUrlSlice {
				select {
				case <-done:
					return
				case bookUrls <- bookUrl:
				}
			}
		}()
		return bookUrls
	}

	// While there are more genre urls to grab ...
	for true {

		// Grab genre url from the genre urls channel.	
		genreUrl, more := <- genreUrls

		// If worker gets a genre url ...
		if more {
			
			genre := strings.Split(strings.Split(genreUrl, "/")[0], "_")[0]
			bookUrlSlice := getBookUrls(genreUrl) // Scrape genre url for the links to the individual book detail webpages.

			// Put bookUrls in a channel for scraping by goroutines.
			done := make(chan interface{})
			defer close(done)
			bookUrls := bookUrlStream(done, bookUrlSlice)

			// Create a slice of stealing queues.
			qs := queue.NewQueueSlice(threads)
			availableToSteal := make([]bool, threads)
			var nothingLeftToStealCount int64

			// Spawn more goroutines to scrape data from urls in bookUrlStream
			for i := 0; i < threads; i ++ {
				detailWorkerCounter += 1
				availableToSteal[i] = true
				go bookDetailWorker(threads, bookDict, genre, bookUrls, doneBookDetailWorker, qs, i, availableToSteal, &nothingLeftToStealCount)
			}

		} else { // Otherwise, if there are no more genre urls, then goroutine worker exits.

			// Wait for all book details to be scraped.
			for i := 0; i < detailWorkerCounter; i++ {
				<- doneBookDetailWorker
			}

			// Indicate that the worker is done and return.
			doneWorker <- true 
			return
		}
	}
}

// main determines whether the sequential or parallel version of the program should be run based on the command line argument.
// main spawns a goroutine to read in standard input and puts read in urls in a channel for consumption by workers.
// main spawns n worker goroutines to consume genre urls for scraping of data.
func main() {

	if len(os.Args) < 1 || len(os.Args) > 2 { // If invalid arguments then print usage statement.

		printUsage()

	} else if len(os.Args) == 2 { // Run parallel version.

		// Title of output document
		fmt.Println("BEST BOOKS TO GET BY GENRE\n")

		// Make global variables for goroutines.
		threads, _ := strconv.Atoi(strings.Split(os.Args[1],"=")[1])
		runtime.GOMAXPROCS(threads)
		bookDict := linkedlist.NewList()

		// Create channels for synchronization of goroutines.
		genreUrls := make(chan string)
		doneWorker := make(chan bool)

		// Read from standard input (initial generator).
		// Genre urls go to genreUrls channel which is consumed by multiple workers (fan out).
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				genreUrl := scanner.Text()
				genreUrls <- genreUrl
			}
			close(genreUrls)
		}()

		// Spawn worker goroutines to scrape all individual book pages from a given genre inde webpage.
		for i := 0; i < threads; i++ {
			go worker(threads, doneWorker, genreUrls, bookDict)
		}

		// Wait for all workers to return.
		for i := 0; i < threads; i++ {
			<-doneWorker
		}

		// Print book dict.
		bookDict.Print()

	} else { // If no parallel flag specified then run the sequential version.

		// Title of output document.
		fmt.Println("BEST BOOKS TO GET BY GENRE\n")

		// Initiate regular Golang dictionary.
		bookDict := make(map[string]*book.Book)

		// Read in image tasks from standard input.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			// Get book index urls from given genre page.
			endOfUrl := scanner.Text()
			bookUrls := getBookUrls(endOfUrl)
			genre := strings.Split(strings.Split(endOfUrl, "/")[0], "_")[0]

			// Get book attributes from book index pages.
			bookBase := "http://books.toscrape.com/catalogue/"
			for _, bookUrl := range bookUrls {

				// Find book details.
				bookUrl = bookBase + bookUrl[9:]
				bookUrlResponse := html.GetBodyResponse(bookUrl)
				book := html.GetBookDetails(bookUrlResponse)

				// If book is cheaper than current front runner then replace.
				b, ok := bookDict[genre]
				if !ok {
					bookDict[genre] = book
				} else if (book.Price < b.Price) {
					bookDict[genre] = book
				}
			}
		}

		// Save best book details in each genre to standard output.
		for key, _ := range bookDict {
			// Cleaning of in stock attribute for export.
			var inStockString string
			if bookDict[key].InStock {
				inStockString = "Yes"
			} else {
				inStockString = "No"
			}
			// Print to std out.
			fmt.Println(strings.Trim(strings.ToUpper(key), " "), "\nTitle:", bookDict[key].Title, "\nDescription:", bookDict[key].Description, "\nPrice:", fmt.Sprintf("%.2f", bookDict[key].Price), "Pound Sterling\nIn Stock?:", inStockString, "\nStars:", bookDict[key].Stars, "\n")			
		}
	}
}