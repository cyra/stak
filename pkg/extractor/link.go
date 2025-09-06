package extractor

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type LinkExtractor struct {
	client *http.Client
	urlRegex *regexp.Regexp
}

func NewLinkExtractor() *LinkExtractor {
	return &LinkExtractor{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		urlRegex: regexp.MustCompile(`https?://[^\s]+`),
	}
}

func (le *LinkExtractor) ExtractLinks(content string) []string {
	return le.urlRegex.FindAllString(content, -1)
}

func (le *LinkExtractor) GetURLTitle(url string) (string, error) {
	resp, err := le.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	title := le.findTitle(doc)
	if title == "" {
		return extractDomain(url), nil
	}

	title = strings.TrimSpace(title)
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	return title, nil
}

func (le *LinkExtractor) findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			return n.FirstChild.Data
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := le.findTitle(c); title != "" {
			return title
		}
	}

	return ""
}

func extractDomain(url string) string {
	domainRegex := regexp.MustCompile(`https?://([^/]+)`)
	matches := domainRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		domain := matches[1]
		if strings.HasPrefix(domain, "www.") {
			domain = domain[4:]
		}
		return domain
	}
	return "Link"
}