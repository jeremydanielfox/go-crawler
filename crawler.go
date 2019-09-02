package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (bool, string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			return true, a.Val
		}
	}
	return false, ""
}

type webpage struct {
	tokens *html.Tokenizer
	host   string
}

// Get all the links on the page that stay within the domain.
func getLinks(page webpage) []string {
	var urls []string
	for {
		tt := page.tokens.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return urls
		case tt == html.StartTagToken:
			t := page.tokens.Token()
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			// Extract the href value, if there is one
			ok, childPath := getHref(t)
			if !ok {
				continue
			}
			// Checking string containment is not really good enough.
			// We should be doing URL regex. Now that I have acknowledged
			// this, let's continue doing string containment :)
			if _, err := url.Parse(childPath); err == nil && strings.Contains(childPath, page.host) {
				// Legitimate sub url
				urls = append(urls, childPath)
			}
		}
	}
}

func main() {
	worklist := make(chan string)
	unseenLinks := make(chan string)
	// Grab the URL from command line, assuming it's the first arg.
	mainURL := os.Args[1:][0]
	// Parse the URL here so we can match on the hostname later.
	u, err := url.Parse(mainURL)
	if err != nil {
		panic(err)
	}
	go func() { worklist <- mainURL }()
	for i := 0; i < 10; i++ {
		go func() {
			for currURL := range unseenLinks {
				fmt.Println(currURL)
				if _, err := url.Parse(currURL); err != nil {
					fmt.Printf("Could not parse as valid URL: %s", currURL)
					continue
				}
				res, err := http.Get(currURL)
				if err != nil {
					panic(err)
				}
				defer res.Body.Close()
				z := html.NewTokenizer(res.Body)
				links := getLinks(webpage{
					z, u.Host,
				})
				for _, link := range links {
					// Doesn't work unless it's in a goroutine. Why?
					go func() { worklist <- link }()
				}
			}
		}()
	}

	seen := make(map[string]bool)
	for link := range worklist {
		if !seen[link] {
			seen[link] = true
			unseenLinks <- link
		}
	}
}
