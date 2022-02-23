package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

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

func main() {
	if len(os.Args) != 2 {
		log.Fatal("You have to provide a root url to start crawling")
	}
	baseurl := os.Args[1]
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
	go func() { urls <- os.Args[1:] }()

	alreadyScanned := make(map[string]bool)
	for ; n > 0; n-- {
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
}
