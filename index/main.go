package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	linkref "src/linkCrawler"
	"strings"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

// type loc struct {
// 	Value string `xml:"loc"`
// }

// type urlset struct {
// 	Urls  []loc  `xml:"url"`
// 	Xmlns string `xml:"xmlns,attr"`
// }

func main() {
	urlFlag := flag.String("url", "https://www.amazon.com/", "the url that you want to build sitemap for")
	maxDepth := flag.Int("depth", 3, "the maximum number of links deep to traverse")
	flag.Parse()
	pages := bfs(*urlFlag, *maxDepth)
	// toXML := urlset{
	// 	Urls:  make([]loc, len(pages)),
	// 	Xmlns: xmlns,
	// }
	f, err := os.Create("newCsv.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(pages); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

type empty struct{}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]empty)
	var q map[string]empty
	nq := map[string]empty{
		urlStr: empty{},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]empty)
		if len(q) == 0 {
			break
		}
		for url, _ := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = empty{}
			for _, link := range get(url) {
				if _, ok := seen[link]; !ok {
					nq[link] = empty{}
				}
			}
		}
	}
	var ret = make([]string, 0, len(seen))
	for url, _ := range seen {
		ret = append(ret, url)
	}
	return ret
}

func get(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
	reqURL := resp.Request.URL
	baseURL := &url.URL{
		Scheme: reqURL.Scheme,
		Host:   reqURL.Host,
	}

	base := baseURL.String()
	return filter(hrefs(resp.Body, base), withPrefix(base))
}

func hrefs(r io.Reader, base string) []string {
	links, _ := linkref.Parse(r)
	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}
	return ret
}

func filter(links []string, keepFn func(string) bool) []string {
	var ret []string
	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}
	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
