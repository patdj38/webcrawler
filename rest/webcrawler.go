package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pat/urls"
	"strings"
)

var Ongoing = false

var scan_token = make(chan struct{}, 20)

func (s *Server) webCrawler(w http.ResponseWriter, r *http.Request) {
	var badRequestError = &Error{ID: "bad_request", Status: 400, Title: "Bad request", Detail: "No based url specified."}
	var badURLError = &Error{ID: "bad_request", Status: 400, Title: "Bad request", Detail: "Provided url can't be parsed."}
	var ServiceUnavailableError = &Error{ID: "bad_request", Status: 503, Title: "Service Unvailable", Detail: "The web crawler is already running."}

	if Ongoing == true {
		writeError(w, r, ServiceUnavailableError)
		return
	}

	//  avoid  concurrent call to the server  (design choice)
	Ongoing = true

	var baseurl string
	if value, ok := r.URL.Query()["url"]; ok {
		baseurl = strings.TrimSpace(value[0])
	} else {
		log.Printf("No based url specified.\n")
		writeError(w, r, badRequestError)
		Ongoing = false
		return
	}

	base, err := url.Parse(baseurl)
	if err != nil {
		log.Printf(err.Error() + "\n")
		badRequestError.Detail = err.Error()
		writeError(w, r, badURLError)
		Ongoing = false
		return
	}
	var root urls.Node
	root.Url = baseurl
	//root.Parent = nil
	urlslist := make(chan []*urls.Urlsstore)
	var n int = 1
	//give arg[1:]  as a list of Urls (will contains only args[1] at least in a first time

	iurl := &urls.Urlsstore{Url: baseurl, Node: &root, Skip: false}
	go func() { urlslist <- []*urls.Urlsstore{iurl} }()

	alreadyScanned := make(map[string]bool)
	for ; n > 0; n-- {
		//fmt.Printf("Pat Add for loop with n=%d \n", n)

		list := <-urlslist
		//fmt.Printf("Pat Add in s.webCrawler list n is %d and list is  : ", n)
		for _, i := range list {
			fmt.Printf("%p %s | ", i.Node, i.Url)
		}
		fmt.Printf("\n")

		for _, url := range list {
			//	fmt.Printf("Pat Add for url loop with url=%+v and list length %d\n", *url, len(list))

			if !alreadyScanned[url.Url] {
				n++
				alreadyScanned[url.Url] = true
				go func(url *urls.Urlsstore) {

					urlslist <- scanUrl(url, base)
				}(url)
			}
		}
	}
	urls.PrintNode(&root, 0)
	encodeJSONResponse(w, r, root)
	urls.VisitedURLs = make(map[string]bool)
	Ongoing = false
}

func scanUrl(myurl *urls.Urlsstore, baseurl *url.URL) []*urls.Urlsstore {
	scan_token <- struct{}{} // we wait to acquire a token
	fmt.Printf("Pat Add in scanUrl url.Node=%p and url %s\n", myurl.Node, myurl.Url)
	list, err := urls.Sitemap(myurl, baseurl)
	<-scan_token // no it's time to release our token

	if err != nil {
		log.Print(err)
	}
	/*	fmt.Printf("Pat Add We are about to return the following list :")
		for _, i := range list {
			fmt.Printf("%p %s | ", i.Node, i.Url)
		}
		fmt.Printf("\n")*/
	return list
}
