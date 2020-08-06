package linkpreview

import (
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	hbot "github.com/whyrusleeping/hellabot"

	logger "gopkg.in/inconshreveable/log15.v2"
)

var lgr = logger.Root()

var linkPreviewRegex = regexp.MustCompile(`(?mi)https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
var LinkPreviewTrigger = hbot.Trigger{
	func(b *hbot.Bot, m *hbot.Message) bool {
		if m.Command == "PART" || m.Command == "QUIT" {
			return false
		}
		if m.From == b.Nick || m.To == b.Nick {
			return false
		}

		return linkPreviewRegex.MatchString(m.Content)

	},
	func(b *hbot.Bot, m *hbot.Message) bool {
		r := linkPreviewRegex.FindAllString(m.Content, -1)
		lgr.Debug("Found links for linkpreview", "url", r)
		for p := range r {
			if p > 2 {
				break
			}
			pageData := fetchContents(r[p])
			if len(pageData) == 0 {
				lgr.Debug("No content from url", "url", p)
				return false
			}
			title := getTitle(pageData)
			if len(title) >= 1 {
				lgr.Info("Found title for URL", "title", title, "url", p)
				b.Reply(m, title)
			} else {
				lgr.Info("Found no title for URL", "url", p)
			}
		}
		return false
	},
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func fetchContents(url string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Buttsbot link-previews")

	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	pageContent := string(respBytes)

	return pageContent
}

func getTitle(s string) string {
	titleStartIndex := strings.Index(s, "<title>")
	if titleStartIndex == -1 {
		return ""
	}
	// Skip to end of title declaration
	titleStartIndex += 7

	titleEndIndex := strings.Index(s, "</title>")
	if titleEndIndex == -1 {
		return ""
	}

	title := s[titleStartIndex:titleEndIndex]
	title = html.UnescapeString(title)
	title = strings.Replace(title, "\n", " - ", -1)
	title = strings.TrimSpace(title)

	if len(title) > 80 {
		title = title[:80] + "..."
	}

	return title
}
