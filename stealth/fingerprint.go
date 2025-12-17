package stealth

import (
	"math/rand"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"

	"linkedin-automation/config"
)

// ApplyFingerprint randomizes UA/viewport and disables webdriver flags.
func ApplyFingerprint(page *rod.Page, cfg config.Config) {
	// Apply rod's stealth patches at the browser level if possible.
	if page.Browser() != nil {
		_, _ = stealth.Page(page.Browser())
	}

	if len(cfg.Browser.UserAgents) > 0 {
		ua := cfg.Browser.UserAgents[rand.Intn(len(cfg.Browser.UserAgents))]
		_ = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: ua})
	}

	if len(cfg.Browser.Viewports) > 0 {
		vp := cfg.Browser.Viewports[rand.Intn(len(cfg.Browser.Viewports))]
		_ = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:             int(vp.Width),
			Height:            int(vp.Height),
			DeviceScaleFactor: vp.Scale,
			Mobile:            false,
		})
	}

	if cfg.Browser.Timezone != "" {
		_ = proto.EmulationSetTimezoneOverride{
			TimezoneID: cfg.Browser.Timezone,
		}.Call(page)
	}

	// Best-effort navigator.webdriver false.
	_, _ = page.Eval(`() => { Object.defineProperty(navigator, 'webdriver', {get: () => false}); }`)
}

