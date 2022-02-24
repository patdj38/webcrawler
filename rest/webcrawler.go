package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pat/urls"
)

var scan_token = make(chan struct{}, 20)

func printNode(node *urls.Node, n int) {

	for i := 0; i < n; i++ {
		fmt.Print("  ")
	}
	if n > 0 {
		fmt.Print("- ")
	}
	fmt.Println(node.Url)
	for _, c := range node.Children {
		printNode(c, n+1)
	}
}

func (s *Server) webCrawler(w http.ResponseWriter, r *http.Request) {

	baseurl := "https://www.google.com"
	baseurl = "https://www.google.com/doodles/royal-wedding"
	//	fmt.Printf("Pat Add on est la \n")
	base, err := url.Parse(baseurl)
	if err != nil {
		log.Fatal(err)
	}
	var root urls.Node
	root.Url = baseurl

	//fmt.Printf("Pat Add %s\n", root)
	urls := make(chan []string)
	var n int = 1

	//give arg[1:]  as a list of Urls (will contains only args[1] at least in a first time
	go func() { urls <- []string{baseurl} }()

	alreadyScanned := make(map[string]bool)
	for ; n > 0; n-- {
		//		fmt.Printf("Pat Add puis on est la \n")

		list := <-urls
		for _, url := range list {

			if !alreadyScanned[url] {
				n++
				alreadyScanned[url] = true
				go func(url string) {

					urls <- scanUrl(url, base, &root)
				}(url)
			}
		}
	}

	encodeJSONResponse(w, r, root)
}

func scanUrl(myurl string, baseurl *url.URL, node *urls.Node) []string {
	scan_token <- struct{}{} // we wait to acquire a token
	fmt.Printf("Pat Add in scanUrl url=%s\n", myurl)
	list, err := urls.Sitemap(myurl, baseurl, node)
	<-scan_token // no it's time to release our token

	if err != nil {
		log.Print(err)
	}
	printNode(node, 0)
	return list
}
