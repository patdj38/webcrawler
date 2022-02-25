package urls

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"pat/utils"
	"path"
	"regexp"
	"sync"

	"golang.org/x/net/html"
)

// this is the struct u(chained data) that is used to store the whole sitemap
// Url is the corresponding url and Children is an array of Node which represent
// the current Node's children

type Node struct {
	Url      string
	Children []*Node
}

// this is the struct used to described a URL that wil be stored for the sitemap
// Url is a string representing the url, Node is a pointer to the corresponding Node
// which represent that URL and skip is a boolean that is used to indicate if this URL should
// be followed (skip=false) or not

type Urlsstore struct {
	Url  string
	Node *Node
	Skip bool
}

// mutex to control access to map
var access sync.Mutex

// regexp to detect link with parameter and anchor I chose to not follow that link in the web crawler mechanism.
// It could be a configurable parameter
var q_regexp = regexp.MustCompile(`.*[\?#].*`)

// store Urls already visited to avoid cycles (reseted after each execution)
var VisitedURLs = make(map[string]bool)

// this function aims to build a sitemap for  the url passed in argument (url.URL field)
// in collect all links of the considered page (make some checks) and return these links as a list of URLs
// to be treated as well. if links not strat with the based url pattern (e.g  the root url we provided) then it
// is considered as a leaf and we not follow this link (ask in the exercise)

func Sitemap(url *Urlsstore, baseurl *url.URL) ([]*Urlsstore, error) {
	fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s\n", url.Node, url.Url)

	var links []*Urlsstore

	// check if the url has the expected form http[s]://
	// that's a design choice. We could refine this check
	if !utils.IsCorrrectURL(url.Url) {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because urlpattern doesn't match\n", url.Node, url.Url)
		return links, nil
	}

	// check if the url is under the initial base url
	if !utils.IsUnderBaseUrl(baseurl.String(), url.Url) {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because basepattern doesn't match\n", url.Node, url.Url)
		return links, nil
	}

	// get the content of the page for teh considered url
	resp, err := http.Get(url.Url)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because http get failed  %s\n", url.Node, url.Url, err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because http status <>200 %d  %s\n", url.Node, url.Url, resp.StatusCode)
		return nil, fmt.Errorf("Unexpected return code when trying to get url: %s (%d)", url, resp.Status)
	}

	// We don't try follow links with mime type <> text/html
	// It's a choice and is could maybe be refined
	if !utils.IsTextURL(resp) {
		//	fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because url is not a text url \n", url.Node, url.Url)
		return nil, nil
	}

	access.Lock() // take a lock to avoid concurrent read/write on the map
	if !VisitedURLs[url.Url] {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we add to already seen  \n", url.Node, url.Url)
		VisitedURLs[url.Url] = true
	} else {
		fmt.Printf("Pat Add in SiteMap url.Node=%p and url=%s we return because url already seen \n", url.Node, url.Url)
		return nil, nil
	}
	access.Unlock()

	//  get and return all links of that page
	list := getLinks(resp.Body, url, baseurl)

	return list, nil
}

func getLinks(body io.Reader, parent *Urlsstore, baseUrl *url.URL) []*Urlsstore {
	fmt.Printf("Pat Add in getlinks url.Node=%p and URL=%s\n", parent.Node, parent.Url)

	var links []*Urlsstore
	z := html.NewTokenizer(body)
	for {
		token := z.Next()

		switch token {
		case html.ErrorToken:
			//fmt.Printf("Pat Add from getLinks we return the following list %+v\n", links)
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
						lta := baseUrl.ResolveReference(u).String() // we try to get the full url
						access.Lock()
						if !VisitedURLs[lta] {
							child := Node{Url: lta}
							parent.Node.Children = append(parent.Node.Children, &child)
							us := &Urlsstore{Url: lta, Node: &child, Skip: false}
							if q_regexp.MatchString(path.Base(lta)) {
								us.Skip = true
							} else {
								us.Skip = false
							}
							links = append(links, us)
							//		fmt.Printf("Pat Add link is %s, parent is %s\n", lta, parent)
						}
						access.Unlock()
					}
				}
			}

		}
	}
	return links
}

func PrintNode(node *Node, n int) {

	for i := 0; i < n; i++ {
		fmt.Print("  ")
	}
	if n > 0 {
		fmt.Print("- ")
	}
	fmt.Println(node.Url)
	for _, c := range node.Children {
		PrintNode(c, n+1)
	}
}
