package browser

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"

	"linkedin-automation/config"
)

// Session wraps a configured Rod browser and page.
type Session struct {
	Browser *rod.Browser
	Page    *rod.Page
	cfg     config.Config
}

// NewSession launches a hardened browser session with human-like defaults.
func NewSession(ctx context.Context, cfg config.Config) (*Session, error) {
	l := launcher.New().
		Headless(cfg.Headless).
		Set("disable-blink-features", "AutomationControlled").
		Set("no-sandbox")

	if cfg.Browser.Proxy != "" {
		l = l.Proxy(cfg.Browser.Proxy)
	}

	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if cfg.Debug {
		browser = browser.Trace(true).SlowMotion(time.Duration(rand.Intn(200)+50) * time.Millisecond)
	}

	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: cfg.BaseURL})
	if err != nil {
		return nil, fmt.Errorf("new page: %w", err)
	}

	// Apply stealth patches at browser level
	_, _ = stealth.Page(browser)

	// Set timezone if configured
	if cfg.Browser.Timezone != "" {
		_ = proto.EmulationSetTimezoneOverride{
			TimezoneID: cfg.Browser.Timezone,
		}.Call(page)
	}

	// Set user agent if configured (pick random from pool if available)
	if len(cfg.Browser.UserAgents) > 0 {
		ua := cfg.Browser.UserAgents[rand.Intn(len(cfg.Browser.UserAgents))]
		if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: ua,
		}); err != nil {
			log.Printf("warn: set user agent: %v", err)
		}
	}

	// Set viewport if configured (pick random from pool if available)
	if len(cfg.Browser.Viewports) > 0 {
		vp := cfg.Browser.Viewports[rand.Intn(len(cfg.Browser.Viewports))]
		if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:             int(vp.Width),
			Height:            int(vp.Height),
			DeviceScaleFactor: vp.Scale,
			Mobile:            false,
		}); err != nil {
			log.Printf("warn: set viewport: %v", err)
		}
	}

	return &Session{
		Browser: browser,
		Page:    page,
		cfg:     cfg,
	}, nil
}

// Close gracefully closes the session.
func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.Page != nil {
		_ = s.Page.Close()
	}
	if s.Browser != nil {
		_ = s.Browser.Close()
	}
}
