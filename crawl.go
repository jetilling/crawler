package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"test_projects/crawler/dataStore"

	_ "github.com/lib/pq"
	"golang.org/x/net/html"
)

func main() {

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=crawler sslmode=disable")

	if err != nil {
		panic(err)
	}
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	dataStore.InitStore(&dataStore.DBStore{DB: db})
	lastLinkRetrieved, err := dataStore.Store.RetrieveLastUsedLink()
	if err != nil {
		WriteErrorToFile(err.Error())
	}
	if len(lastLinkRetrieved) == 0 {
		lastLinkRetrieved = "https://en.wikipedia.org/wiki/Main_Page"
	}

	fmt.Println(" ")
	fmt.Println("/*****************************************************/")
	fmt.Println("  STARTING CRAWL!")
	fmt.Println("  Initiating with site: ", lastLinkRetrieved)
	fmt.Println("/*****************************************************/")
	fmt.Println(" ")

	queue := make(chan string)

	go func() { queue <- lastLinkRetrieved }()

	for uri := range queue {
		if !strings.Contains(uri, "twitter") {
			if !strings.Contains(uri, "facebook") {
				time.Sleep(2 * time.Second)
				enqueue(uri, queue)
			}
		}
	}
}

func enqueue(uri string, queue chan string) {
	fmt.Println("fetching", uri)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)
	if err != nil {
		// panic(err)
		fmt.Println(err)
		WriteErrorToFile(err.Error())
		return
		// write the error to a file and move on
		// maybe write failed url to database in failed_links table??
	}

	defer resp.Body.Close()

	// fmt.Println("adding: ", uri)
	// dataStore.Store.AddLink(uri, resp.StatusCode)

	links := All(resp.Body, uri, resp.StatusCode)

	for _, link := range links {
		absolute := fixUrl(link, uri)
		if uri != "" {
			go func() { queue <- absolute }()
		}
	}
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		panic(err)
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}

// All takes a reader object (like the one returned from http.Get())
// It returns a slice of strings representing the "href" attributes from
// anchor links found in the provided html.
// It does not close the reader passed to it.
// https://drstearns.github.io/tutorials/tokenizing/
func All(httpBody io.Reader, uri string, statusCode int) []string {
	links := []string{}
	col := []string{}
	page := html.NewTokenizer(httpBody)
	for {
		tokenType := page.Next()

		if tokenType == html.ErrorToken {
			return links
		}
		if tokenType == html.StartTagToken {
			token := page.Token()

			switch token.DataAtom.String() {
			// case "h1":
			// 	//the next token should be the page title
			// 	tokenType = page.Next()
			// 	//just make sure it's actually a text token
			// 	if tokenType == html.TextToken {
			// 		//report the page title and break out of the loop
			// 		dataStore.Store.AddLink(uri, statusCode, page.Token().Data)
			// 	}
			case "title":
				//the next token should be the page title
				tokenType = page.Next()
				//just make sure it's actually a text token
				if tokenType == html.TextToken {
					//report the page title and break out of the loop
					dataStore.Store.AddLink(uri, statusCode, page.Token().Data)
				}
			}

			if token.DataAtom.String() == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						tl := trimHash(attr.Val)
						col = append(col, tl)
						resolv(&links, col)
					}
				}
			}
		}
		// if len(token.DataAtom.String()) > 0 {
		// 	fmt.Println("TOKEN: ", token)
		// }
		// if token.DataAtom.String() == "h1" {
		// 	fmt.Println("TOKEN: ", page.Text())
		// }
	}
}

// trimHash slices a hash # from the link
func trimHash(l string) string {
	if strings.Contains(l, "#") {
		var index int
		for n, str := range l {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return l[:index]
	}
	return l
}

// check looks to see if a url exits in the slice.
func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

// resolv adds links to the link slice and insures that there is no repetition
// in our collection.
func resolv(sl *[]string, ml []string) {
	for _, str := range ml {
		if check(*sl, str) == false {
			*sl = append(*sl, str)
		}
	}
}
