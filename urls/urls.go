package urls

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/net/html"
)

type Node struct {
	Url      string
	Children []*Node
}

var full_url_map = make(map[string]string)

func Sitemap(url string, baseurl *url.URL, node *Node) ([]string, error) {
	fmt.Printf("Pat Add in SiteMap url=%s\n", url)

	var links []string
	urlpattern := regexp.MustCompile("^https?://.*")
	if !urlpattern.MatchString(url) {
		return links, nil
	}

	basepattern := regexp.MustCompile("^" + baseurl.String() + "/?.*")
	if !basepattern.MatchString(url) {
		return links, nil
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	if _, ok := full_url_map[url]; !ok {
		full_url_map[url] = ""
	}
	list := getLinks(resp.Body, url, baseurl, node)

	for i, c := range list {
		Sitemap(c, baseurl, node.Children[i])
	}
	return nil, nil
}

func getLinks(body io.Reader, parent string, baseUrl *url.URL, node *Node) []string {
	fmt.Printf("Pat Add in getlinks url=%s\n", baseUrl)

	var links []string
	z := html.NewTokenizer(body)
	for {
		token := z.Next()

		switch token {
		case html.ErrorToken:
			//todo: urls list shoudn't contain duplicates
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, a := range token.Attr {
					if a.Key == "href" { // we are interested only by http links
						u, err := url.Parse(a.Val)
						if err != nil || u == nil {
							log.Printf("Error parsing url %s\n", a.Val)
							continue
						}
						lta := baseUrl.ResolveReference(u).String() // we get the full url
						if _, ok := full_url_map[lta]; !ok {
							full_url_map[lta] = ""
							child := Node{Url: lta}
							node.Children = append(node.Children, &child)
							links = append(links, lta)
							//		fmt.Printf("Pat Add link is %s, parent is %s\n", lta, parent)
						}
					}
				}
			}

		}
	}
	return links
}
