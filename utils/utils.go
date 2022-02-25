package utils

import (
	"net/http"
	"regexp"
)

var TextUrlPattern = regexp.MustCompile(`text/html.*`)
var URLPattern = regexp.MustCompile("^https?://.*")

func MatchString(re *regexp.Regexp, s string) bool {
	return re.MatchString(s)
}

func IsTextURL(resp *http.Response) bool {
	return MatchString(TextUrlPattern, resp.Header.Get("Content-Type"))
}

func IsCorrrectURL(url string) bool {
	return MatchString(URLPattern, url)
}

func IsUnderBaseUrl(base string, url string) bool {
	basepattern := regexp.MustCompile("^" + base + "/?.*")
	return MatchString(basepattern, url)
}
