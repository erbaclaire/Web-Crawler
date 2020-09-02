package html

import (
	"src/book"
	"log"
	"net/http"
	"golang.org/x/net/html"
	// "fmt"
	"io"
	"strconv"
	"strings"
)

// GetBodyResponse returns an io.Reader of the url's html body response.
func GetBodyResponse(url string) io.ReadCloser {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	return res.Body
}

// ParseHtmlForBookLinks traverses the tree like structure of the html of 
// the webpage and finds the links to the individual book's index page.
func ParseHtmlForBookLinks(n *html.Node, bookUrls *[]string) {
	if n.Type == html.ElementNode && n.Data == "h3" {
		*bookUrls = append(*bookUrls, n.FirstChild.Attr[0].Val)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ParseHtmlForBookLinks(c, bookUrls)
	}
}

// GetBookUrls aggregates indiviudal books' index page urls in to a slice
// so that those urls can be crawled, as well.
func GetBookUrls(r io.ReadCloser, rString string) []string {
	var bookUrls []string
	for { 
		doc, err := html.Parse(r)
		if err != nil {
			log.Fatal(err)
		}
		ParseHtmlForBookLinks(doc, &bookUrls)
		tmpNext := "no next"
		next := Next(doc, &tmpNext) // Find the next page.
		if next == "no next" { // If no next page then have all individual book urls for category.
			break
		} else { // Otherwise get body response from the next page.
			rStringSplit := strings.Split(rString, "/")
			baseUrl := strings.Join(rStringSplit[:len(rStringSplit)-1], "/")
			r = GetBodyResponse(baseUrl + "/" + next)
		}
	}
	return bookUrls
}

// Next grabs the next page link if there is one.
func Next(n *html.Node, tmpNext *string) string {
	if n.Type == html.ElementNode && n.Data == "li"  && len(n.FirstChild.Attr) > 0 && len(n.Attr) > 0 && n.Attr[0].Val == "next" {
		*tmpNext = n.FirstChild.Attr[0].Val
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		Next(c, tmpNext)
	}
	return *tmpNext
}

// ParseHtmlForBookDetails searches for the book's title, price, if in stock or not, and stars.
func ParseHtmlForBookDetails(n *html.Node, book *book.Book, titleText bool, priceText bool) {
    
    // Set book title.
    if titleText && n.Type == html.TextNode {
        book.Title = n.Data
    }

    // Set description
    if n.Type == html.ElementNode && n.Data == "meta" && len(n.Attr) > 0 && n.Attr[0].Val == "description" {
    	desc := n.Attr[1].Val
    	book.Description = desc[:len(desc)-1]
    }

    // Set book price.
    if priceText && n.Type == html.TextNode {
    	book.Price, _ = strconv.ParseFloat(n.Data[2:], 64)
    }

    // Set book number of stars.
    if n.Type == html.ElementNode && n.Data == "p" && len(n.Attr) > 0 && strings.Contains(n.Attr[0].Val, "star-rating") && n.Parent.Attr[0].Val == "col-sm-6 product_main" {
    	// fmt.Println(n.Attr[0].Val)
    	starString := strings.Split(n.Attr[0].Val, " ")[1]
    	if starString == "One" {
    		book.Stars = 1
    	} else if starString == "Two" {
    		book.Stars = 2
    	} else if starString == "Three" {
    		book.Stars = 3
    	} else if starString == "Four" {
    		book.Stars = 4
    	} else if starString == "Five" {
    		book.Stars = 5
    	} else {
    		book.Stars = 0
    	}
    }

    // Set availability attribute.
    if n.Type == html.ElementNode && n.Data == "i" && n.Parent.Attr[0].Val == "instock availability" && n.Parent.Parent.Attr[0].Val == "col-sm-6 product_main" {
    	if n.Attr[0].Val == "icon-ok" {
    		book.InStock = true
    	} else {
    		book.InStock = false
    	}
    }

    // Prepare for recursive call.
    titleText = titleText || (n.Type == html.ElementNode && n.Data == "h1") 
    priceText = priceText || (n.Type == html.ElementNode && n.Data == "p" && len(n.Attr) > 0 && n.Attr[0].Val == "price_color" && n.Parent.Attr[0].Val == "col-sm-6 product_main")
    
    // Recursive calls.
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        ParseHtmlForBookDetails(c, book, titleText, priceText)
    }
}

// GetBookDetails calls ParseHtmlForBookDetails toparse the html for the specified attributes.
func GetBookDetails(r io.ReadCloser) *book.Book {
	book := book.NewBook("temp", "temp", 0.0, false, 0) // Initialize a temporary book for updating.
	doc, err := html.Parse(r)
	if err != nil {
		log.Fatal(err)
	}
	ParseHtmlForBookDetails(doc, book, false, false)
	return book
}