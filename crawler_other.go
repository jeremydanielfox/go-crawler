package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

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

	go func() { worklist <- "http://slatestarcodex.com"}()
	for i := 0; i < 10; i++ {
		go func(id int) {
			for currURL := range unseenLinks {
				fmt.Println(id)
				fmt.Println(currURL)
				if _, err := url.Parse(currURL); err != nil {
					panic(err)
				}
				res, err := http.Get(currURL)
				if err != nil {
					panic(err)
				}
				defer res.Body.Close()
				z := html.NewTokenizer(res.Body)
				links := getLinks(webpage{
					z, "slatestarcodex",
				})
				for _, link := range links {
					go func() {worklist <- link}
				}
			}
		}(i)
	}

	seen := make(map[string]bool)
	for link := range worklist {
		if !seen[link] {
			seen[link] = true
			unseenLinks <- link
		}
	}
}
