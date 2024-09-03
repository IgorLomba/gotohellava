package ava

import (
	netURL "net/url"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const (
	invalidLoginMsg = "Invalid login, please try again\nInvalid login, please try again\nAVA UFMS - ENSINO"
	moodleSession   = "MoodleSession"
	avaLoginURL     = "https://ava.ufms.br/login/index.php"
)

var linksMap = new(sync.Map)

func Visit(url, username, password string) {
	log.Debugf("url: %s, username: %s, password: %s", url, username, password)

	browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
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

	var sessionCookie string
	for _, cookie := range page.MustCookies() {
		if cookie.Name == moodleSession {
			sessionCookie = cookie.Value
			break
		}
	}

	log.Debugf("session cookie: %s", sessionCookie)
	log.Info("Logged in successfully")

	page = browser.MustPage(url)
	wg := sync.WaitGroup{}

	barEnabled := !log.IsLevelEnabled(log.DebugLevel)

	var bar *progressbar.ProgressBar
	for {
		page.MustReload().MustWaitLoad()
		links := page.MustElements("a.aalink")
		allLinksVisited := true

		if barEnabled {
			if bar == nil {
				bar = progressbar.Default(int64(len(links)))
				bar.Describe("Visiting links")
			} else {
				bar.Describe("Visiting links")
				bar.ChangeMax(len(links))
			}
		}

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
					log.Errorf("error parsing url: %v", err)
					return
				}

				if _, exists := parsedURL.Query()["id"]; exists {
					log.Debugf("link found: %s", href)
					if _, visited := linksMap.Load(href); visited {
						log.Debugf("link already visited: %s", href)
						return
					}

					allLinksVisited = false
					Get(href, sessionCookie)
					linksMap.Store(href, href)
				}

				if barEnabled {
					_ = bar.Add(1)
				}
			}(link)
		}

		wg.Wait()
		if allLinksVisited {
			break
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
