package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"pat/urls"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <server> <url>", os.Args[0])
	}
	host := os.Args[1]
	urltoCrawl := os.Args[2]

	client, err := myClient(host)
	if err != nil {
		log.Fatalln("An error occurred" + err.Error())
	}
	request, err := client.NewRequest("GET", "/webcrawler", nil)
	if err != nil {
		log.Fatalln("An error occurred" + err.Error())
	}
	request.Header.Add("Accept", "application/json")

	params := url.Values{}

	params.Set("url", urltoCrawl)
	request.URL.RawQuery = params.Encode()

	response, err := client.Do(request)
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()

	if err != nil {
		log.Fatalln("An error occurred" + err.Error())
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected error code :  %d", response.StatusCode)
	}
	sitemap := new(urls.Node)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln("An error occurred" + err.Error())
	}

	err = json.Unmarshal(body, &sitemap)
	if err != nil {
		log.Fatalln("An error occurred" + err.Error())
	}
	urls.PrintNode(sitemap, 0)
	fmt.Printf("Exiting main...\n")
}
