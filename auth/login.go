package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation/config"
	"linkedin-automation/logger"
	"linkedin-automation/stealth"
	"linkedin-automation/utils"
)

// Auth handles LinkedIn session creation, authentication, and cookie management.
// It supports cookie-based session reuse and detects security challenges (CAPTCHA, 2FA).
type Auth struct {
	Cfg config.Config
	Log *logger.Logger
}

// StartSession creates a browser session and ensures authentication.
// It attempts to reuse saved cookies first, falling back to login if needed.
// Returns the browser, authenticated page, or an error if authentication fails.
// Detects and stops safely on CAPTCHA, 2FA, or invalid credentials.
func (a *Auth) StartSession(ctx context.Context) (*rod.Browser, *rod.Page, error) {
	browser, page, err := a.newBrowser(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err := a.tryLoadCookies(page); err == nil {
		if a.isLoggedIn(page) {
			a.Log.Info("session restored from cookies")
			return browser, page, nil
		}
	}

	if err := a.login(ctx, page); err != nil {
		return nil, nil, err
	}
	return browser, page, nil
}

func (a *Auth) newBrowser(ctx context.Context) (*rod.Browser, *rod.Page, error) {
	l := launcher.New().Headless(a.Cfg.Headless).Set("disable-blink-features", "AutomationControlled")
	if a.Cfg.Browser.Proxy != "" {
		l = l.Proxy(a.Cfg.Browser.Proxy)
	}
	u, err := l.Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("launch browser: %w", err)
	}
	browser := rod.New().ControlURL(u)
	if a.Cfg.Debug {
		browser = browser.Trace(true)
	}
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("connect browser: %w", err)
	}
	page, err := browser.Page(proto.TargetCreateTarget{URL: a.Cfg.BaseURL})
	if err != nil {
		return nil, nil, fmt.Errorf("open page: %w", err)
	}
	stealth.ApplyFingerprint(page, a.Cfg)
	return browser, page, nil
}

func (a *Auth) tryLoadCookies(page *rod.Page) error {
	if a.Cfg.Paths.CookieFile == "" {
		return errors.New("no cookie file configured")
	}
	var cookies []*proto.NetworkCookie
	if err := utils.LoadJSON(a.Cfg.Paths.CookieFile, &cookies); err != nil {
		return err
	}
	cparams := []*proto.NetworkCookieParam{}
	for _, c := range cookies {
		cparams = append(cparams, &proto.NetworkCookieParam{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			Path:   c.Path,
		})
	}
	return page.SetCookies(cparams)
}

func (a *Auth) isLoggedIn(page *rod.Page) bool {
	info, err := page.Info()
	if err != nil || info == nil || info.URL == "" {
		return false
	}
	if _, err := page.Timeout(8 * time.Second).Element(`input[placeholder="Search"]`); err == nil {
		return true
	}
	return false
}

func (a *Auth) login(ctx context.Context, page *rod.Page) error {
	if err := page.Navigate(a.Cfg.BaseURL + "/login"); err != nil {
		return fmt.Errorf("navigate login: %w", err)
	}

	// Handle "Welcome back" account picker: click "Sign in using another account" if present.
	if other, err := page.Timeout(10 * time.Second).ElementR("*", "Sign in using another account"); err == nil {
		_ = other.Click(proto.InputMouseButtonLeft, 1)
	}

	// Wait for the standard login form; be tolerant of slow loads.
	if _, err := page.Timeout(30 * time.Second).Element(`form`); err != nil {
		return fmt.Errorf("login form did not appear (possible network/selector issue): %w", err)
	}

	emailInput, err := page.Timeout(30 * time.Second).Element(`input#username`)
	if err != nil {
		return fmt.Errorf("email input not found: %w", err)
	}
	passInput, err := page.Timeout(30 * time.Second).Element(`input#password`)
	if err != nil {
		return fmt.Errorf("password input not found: %w", err)
	}

	_ = emailInput.SelectAllText()
	_ = emailInput.Input("")
	if err := typeSlow(emailInput, a.Cfg.Credentials.Email); err != nil {
		return err
	}
	stealth.RandomPause(a.Cfg.Humanization.MinDelayMs, a.Cfg.Humanization.MaxDelayMs, a.Cfg.Humanization.JitterMs)
	_ = passInput.SelectAllText()
	_ = passInput.Input("")
	if err := typeSlow(passInput, a.Cfg.Credentials.Password); err != nil {
		return err
	}

	btn, err := page.Timeout(15 * time.Second).ElementR("button", `^Sign in$`)
	if err != nil {
		return fmt.Errorf("sign in button not found: %w", err)
	}
	box := btn.MustShape().Box()
	_ = stealth.MoveHumanlike(page, box.X+box.Width/2, box.Y+box.Height/2)
	if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click sign in: %w", err)
	}

	// After clicking, poll for logged-in state instead of relying on navigation events.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if a.detectCaptcha(page) {
			a.Log.Warn("captcha detected, stopping")
			return errors.New("captcha detected; please solve it manually and rerun later")
		}
		if a.detectTwoFA(page) {
			a.Log.Warn("2FA prompt detected, manual action required")
			return errors.New("2fa required; complete it manually in the browser")
		}
		if a.detectBadCreds(page) {
			return errors.New("invalid credentials")
		}
		if a.isLoggedIn(page) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !a.isLoggedIn(page) {
		return fmt.Errorf("login did not complete within timeout; please check credentials or 2FA/captcha")
	}

	if a.detectCaptcha(page) {
		a.Log.Warn("captcha detected, stopping")
		return errors.New("captcha detected")
	}
	if a.detectTwoFA(page) {
		a.Log.Warn("2FA prompt detected, manual action required")
		return errors.New("2fa required")
	}
	if a.detectBadCreds(page) {
		return errors.New("invalid credentials")
	}

	if a.Cfg.Paths.CookieFile != "" {
		cookies, _ := page.Cookies([]string{})
		_ = utils.SaveJSON(a.Cfg.Paths.CookieFile, cookies)
	}

	a.Log.Info("login successful")
	return nil
}

func (a *Auth) detectCaptcha(page *rod.Page) bool {
	if _, err := page.Timeout(3 * time.Second).ElementR("div", "captcha"); err == nil {
		return true
	}
	if _, err := page.Timeout(3 * time.Second).Element(`iframe[src*="captcha"]`); err == nil {
		return true
	}
	return false
}

func (a *Auth) detectTwoFA(page *rod.Page) bool {
	if _, err := page.Timeout(3 * time.Second).Element(`input[name="pin"]`); err == nil {
		return true
	}
	return false
}

func (a *Auth) detectBadCreds(page *rod.Page) bool {
	if _, err := page.Timeout(2 * time.Second).ElementR("div", "wrong password"); err == nil {
		return true
	}
	if _, err := page.Timeout(2 * time.Second).ElementR("div", "check your email"); err == nil {
		return true
	}
	return false
}

// typeSlow types text character-by-character with human-like delays but no typos.
func typeSlow(el *rod.Element, text string) error {
	for _, ch := range text {
		if err := el.Input(string(ch)); err != nil {
			return err
		}
		time.Sleep(time.Duration(60+rand.Intn(120)) * time.Millisecond)
	}
	return nil
}

