package ava

import (
	netURL "net/url"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const (
	invalidLoginMsg = "Invalid login, please try again\nInvalid login, please try again\nAVA UFMS - ENSINO"
	moodleSession   = "MoodleSession"
	avaLoginURL     = "https://ava.ufms.br/login/index.php"
)

func Visit(url, username, password string) {
	log.Debugf("url: %s, username: %s, password: %s", url, username, password)

	linksMap := NewLinks()
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

	if strings.Contains(page.MustElement("body").MustText(), invalidLoginMsg) {
		log.Error("Invalid credentials, please try again")
		return
	}

	sessionCookie := ""
	cookies := page.MustCookies()
	for _, cookie := range cookies {
		if cookie.Name == "MoodleSession" {
			sessionCookie = cookie.Value
		}
	}

	log.Debugf("session cookie: %s", sessionCookie)
	log.Info("Logged in successfully")

	page = browser.MustPage(url)

	wg := sync.WaitGroup{}
	for loop {
		page.MustReload().MustWaitLoad()

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
					log.Debugf("error parsing url: %v", err)
					return
				}

				queryParams := parsedURL.Query()
				if _, exists := queryParams["id"]; exists {
					log.Debugf("Link found: %s", href)

					if _, visited := linksMap.Get(href); visited {
						log.Debugf("link already visited: %s", href)
						return
					}

					allLinksVisited = false
					go Get(href, sessionCookie)

					linksMap.Set(href, href)
				}
			}(link)
		}

		wg.Wait()

		if allLinksVisited {
			log.Info("All available links visited")
			loop = false
		}
	}
}

func Get(url, sessionCookie string) {
	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.SetCookie(moodleSession, sessionCookie)

	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)
	if err != nil {
		log.Errorf("error making request: %v", err)
	}
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
}
