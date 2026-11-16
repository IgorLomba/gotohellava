package ava

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

const (
	avaLoginURL    = "https://ava.ufms.br/login/index.php"
	visitTimeout   = 45 * time.Second
	courseTimeout  = 90 * time.Second
	maxAttempts    = 2
	stateDone      = "done"
	stateManual    = "manual"
	stateUntracked = "untracked"
)

type activity struct {
	Name  string `json:"name"`
	Href  string `json:"href"`
	CmID  string `json:"cmid"`
	State string `json:"state"`
}

const collectActivitiesJS = `() => {
	const doneText = /^(conclu[ií]d[oa]|feito|done|completed?)/i;
	const text = (el) => {
		if (!el) return '';
		const clone = el.cloneNode(true);
		clone.querySelectorAll('.accesshide, .sr-only, .visually-hidden').forEach((n) => n.remove());
		return clone.textContent.replace(/\s+/g, ' ').trim();
	};

	let items = [...document.querySelectorAll('li.activity')];
	if (items.length === 0) {
		items = [...document.querySelectorAll('a.aalink')].map((a) => a.closest('.activity-item') || a.parentElement);
	}

	const out = [];
	const seen = new Set();
	for (const item of items) {
		const own = (selector) => [...item.querySelectorAll(selector)].find((el) => {
			const owner = el.closest('li.activity');
			return !owner || owner === item;
		});
		const link = own('a.aalink[href]') || own('.activityname a[href]');
		const href = link ? link.href : '';
		const region = own('[data-region="completion-info"]') || own('.activity-completion, .completion-info');

		let state = 'untracked';
		let cmid = item.dataset.id || '';
		if (region) {
			const manual = region.querySelector('[data-action="toggle-manual-completion"]');
			const button = region.querySelector('button');
			const badges = [...region.querySelectorAll('.badge')];
			if (manual) {
				cmid = manual.dataset.cmid || cmid;
				state = manual.dataset.toggletype === 'manual:undo' ? 'done' : 'manual';
			} else if (button) {
				state = button.classList.contains('btn-success') || doneText.test(text(button)) ? 'done' : 'todo';
			} else if (badges.length > 0) {
				state = badges.every((b) => b.matches('.badge-success, .alert-success, .bg-success')) ? 'done' : 'todo';
			}
		}

		const key = cmid || href;
		if (!key || seen.has(key) || (href && seen.has(href))) continue;
		seen.add(key);
		if (href) seen.add(href);

		out.push({
			name: text(link) || item.dataset.activityname || href,
			href: href,
			cmid: cmid,
			state: state,
		});
	}
	return out;
}`

func Visit(url, username, password string, headed bool) {
	log.Debugf("url: %s, username: %s", url, username)

	browser := rod.New().ControlURL(launcher.New().Headless(!headed).MustLaunch()).MustConnect()
	defer browser.MustClose()

	if err := login(browser, username, password); err != nil {
		log.Error(err)
		return
	}
	log.Info("Logged in successfully")

	course := browser.MustPage("")
	tab := browser.MustPage("")
	acceptDialogs(tab)

	barEnabled := !log.IsLevelEnabled(log.DebugLevel)
	attempts := map[string]int{}
	var last []activity

	for pass := 1; ; pass++ {
		activities, err := loadActivities(course, url)
		if err != nil {
			log.Errorf("error loading the course page: %v", err)
			return
		}
		last = activities

		var pending []activity
		for _, a := range activities {
			if needsVisit(a, attempts) {
				pending = append(pending, a)
			}
		}

		log.Debugf("pass %d: %d activities, %d to visit", pass, len(activities), len(pending))
		if len(pending) == 0 {
			break
		}

		var bar *progressbar.ProgressBar
		if barEnabled {
			bar = progressbar.Default(int64(len(pending)), fmt.Sprintf("Visiting links (round %d)", pass))
		}

		for _, a := range pending {
			attempts[key(a)]++
			log.Debugf("visiting [%s] %s (%s)", a.State, a.Name, a.Href)

			if a.Href != "" {
				if err = open(tab, a.Href); err != nil {
					log.Debugf("could not open %s: %v", a.Href, err)
				}
			}
			if a.State == stateManual {
				if err = markManualDone(course, a); err != nil {
					log.Debugf("could not mark %q as done: %v", a.Name, err)
				}
			}

			if bar != nil {
				_ = bar.Add(1)
			}
		}
	}
	report(last)
}

func login(browser *rod.Browser, username, password string) error {
	page := browser.MustPage(avaLoginURL)
	defer page.MustClose()

	page.MustElement("#username").MustInput(username)
	page.MustElement("#password").MustInput(password)

	wait := page.MustWaitNavigation()
	page.MustElement("#loginbtn").MustClick()
	wait()
	page.MustWaitLoad()

	if strings.Contains(page.MustInfo().URL, "/login/") {
		return errors.New("invalid credentials, please try again")
	}
	return nil
}

func loadActivities(course *rod.Page, url string) ([]activity, error) {
	p := course.Timeout(courseTimeout)
	load := p.Navigate
	if info, err := p.Info(); err == nil && info.URL == url {
		load = func(string) error { return p.Reload() }
	}
	if err := load(url); err != nil {
		return nil, err
	}
	if err := p.WaitLoad(); err != nil {
		return nil, err
	}

	res, err := p.Eval(collectActivitiesJS)
	if err != nil {
		return nil, err
	}

	var activities []activity
	if err = res.Value.Unmarshal(&activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func key(a activity) string {
	if a.CmID != "" {
		return a.CmID
	}
	return a.Href
}

func needsVisit(a activity, attempts map[string]int) bool {
	switch a.State {
	case stateDone:
		return false
	case stateUntracked:
		return a.Href != "" && attempts[key(a)] == 0
	case stateManual:
		return attempts[key(a)] < maxAttempts
	default:
		return a.Href != "" && attempts[key(a)] < maxAttempts
	}
}

func open(tab *rod.Page, href string) error {
	p := tab.Timeout(visitTimeout)

	if err := p.Navigate(href); err != nil {
		var navErr *rod.NavigationError
		if errors.As(err, &navErr) && strings.Contains(navErr.Reason, "ERR_ABORTED") {
			return nil
		}
		return err
	}
	return p.WaitLoad()
}

func markManualDone(course *rod.Page, a activity) error {
	if a.CmID == "" {
		return errors.New("activity has no id")
	}

	selector := fmt.Sprintf(`[data-action="toggle-manual-completion"][data-cmid="%s"]`, a.CmID)
	p := course.Timeout(15 * time.Second)

	button, err := p.Element(selector + `[data-toggletype="manual:mark"]`)
	if err != nil {
		return err
	}
	if _, err = button.Eval(`() => this.click()`); err != nil {
		return err
	}

	_, err = p.Element(selector + `[data-toggletype="manual:undo"]`)
	return err
}

func acceptDialogs(page *rod.Page) {
	go page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		_ = proto.PageHandleJavaScriptDialog{Accept: true}.Call(page)
	})()
}

func report(activities []activity) {
	done, tracked := 0, 0
	var left []activity
	for _, a := range activities {
		switch a.State {
		case stateUntracked:
			continue
		case stateDone:
			done++
		default:
			left = append(left, a)
		}
		tracked++
	}

	fmt.Println()
	log.Infof("%d of %d activities are concluded", done, tracked)
	if len(left) == 0 {
		return
	}

	log.Info("These still need you:")
	for _, a := range left {
		fmt.Printf("  - %s\n    %s\n", a.Name, a.Href)
	}
}
