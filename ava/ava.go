package ava

import (
	"bytes"
	netURL "net/url"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

const (
	moodleSession = "MoodleSession"
	avaLoginURL   = "https://ava.ufms.br/login/index.php"
)

var linksMap = make(map[string]string)

func RodVisit(url, username, password string) {
	loop := true
	l := launcher.New().Headless(true)
	launchURL := l.MustLaunch()
	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(avaLoginURL)

	page.MustElement("#username").MustInput(username)
	page.MustElement("#password").MustInput(password)
	page.MustElement("#loginbtn").MustClick()

	page.MustWaitLoad()

	sessionCookie := ""
	cookies := page.MustCookies()
	for _, cookie := range cookies {
		if cookie.Name == "MoodleSession" {
			sessionCookie = cookie.Value
		}
	}
	page = browser.MustPage(url)

	wg := sync.WaitGroup{}
	for loop {
		page.MustReload()
		page.MustWaitLoad()

		links := page.MustElements("a.aalink")

		allLinksVisited := true
		for _, link := range links {
			wg.Add(1)
			go func(l *rod.Element) {
				defer wg.Done()
				href := l.MustProperty("href").String()
				if href == "" {
					return
				}

				parsedURL, err := netURL.Parse(href)
				if err != nil {
					log.Error().Msgf("error parsing url: %v", err)
					return
				}

				queryParams := parsedURL.Query()
				if _, exists := queryParams["id"]; exists {
					log.Info().Msgf("link found': %s", href)
					if linksMap[href] != "" {
						log.Info().Msgf("link already visited: %s", href)
						return
					}
					allLinksVisited = false
					_ = Get(href, sessionCookie)
					linksMap[href] = href
				}
			}(link)
		}

		wg.Wait()

		if allLinksVisited {
			log.Info().Msg("all links visited")
			loop = false
		}
	}
}

func Get(url, sessionCookie string) string {
	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.SetCookie(moodleSession, sessionCookie)

	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)
	if err != nil {
		log.Error().Msgf("error making request: %v", err)
	}
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	b := bytes.Clone(resp.Body())
	body := string(b)
	return body
}
