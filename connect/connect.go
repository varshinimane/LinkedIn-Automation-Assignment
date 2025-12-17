package connect

import (
	"context"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation/config"
	"linkedin-automation/logger"
	"linkedin-automation/stealth"
	"linkedin-automation/storage"
	"linkedin-automation/utils"
)

// SendConnections visits each profile URL, applies human-like behavior (scrolling, hovering),
// finds and clicks the Connect button, adds personalized notes from templates,
// and enforces daily connection limits via persistent state.
// Skips profiles already contacted and respects business hours if configured.
// Returns an error if the operation fails critically.
func SendConnections(ctx context.Context, page *rod.Page, profiles []string, cfg config.Config, store *storage.Store, log *logger.Logger) error {
	current, _ := store.GetDailyCounter("connections")
	if current >= cfg.Limits.DailyConnections {
		log.Warn("daily connection limit reached", "count", current)
		return nil
	}

	for _, profileURL := range profiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !utils.NowInBusinessHours(cfg.Humanization.BusinessHours.Start, cfg.Humanization.BusinessHours.End) {
			log.Warn("outside business hours, pausing")
			stealth.Cooldown(300)
			continue
		}

		if err := page.Navigate(profileURL); err != nil {
			log.Warn("navigate profile failed", "url", profileURL, "err", err)
			continue
		}

		stealth.RandomScroll(page)
		if shape := page.MustElement(`body`).MustShape().Box(); shape.Width > 0 {
			stealth.WanderCursor(page, shape.Width, shape.Height)
		}

		btn, err := page.Timeout(8 * time.Second).ElementR("button", "Connect")
		if err != nil {
			log.Warn("connect button not found", "url", profileURL)
			continue
		}

		box := btn.MustShape().Box()
		_ = stealth.MoveHumanlike(page, box.X+box.Width/2, box.Y+box.Height/2)
		if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			log.Warn("click connect failed", "url", profileURL, "err", err)
			continue
		}

		time.Sleep(800 * time.Millisecond)
		if noteBtn, err := page.ElementR("button", "Add a note"); err == nil {
			_ = noteBtn.Click(proto.InputMouseButtonLeft, 1)
			time.Sleep(400 * time.Millisecond)
			if textArea, err := page.Element(`textarea[name="message"]`); err == nil {
				name := extractName(page)
				company := extractCompany(page)
				tpl := utils.RandomChoice(cfg.Messaging.Templates)
				body := utils.Template(tpl, map[string]string{
					"name":    name,
					"company": company,
				})
				_ = stealth.TypeHumanWithTypos(textArea, body)
			}
		}

		if sendBtn, err := page.ElementR("button", "Send"); err == nil {
			_ = sendBtn.Click(proto.InputMouseButtonLeft, 1)
		}

		_, _ = store.IncrementDailyCounter("connections")
		_ = store.SaveConnection(profileURL, extractName(page), extractCompany(page), "requested")
		log.Info("connection sent", "url", profileURL)

		if c, _ := store.GetDailyCounter("connections"); c >= cfg.Limits.DailyConnections {
			log.Warn("daily connection limit reached", "count", c)
			return nil
		}

		stealth.RandomPause(cfg.Humanization.MinDelayMs, cfg.Humanization.MaxDelayMs, cfg.Humanization.JitterMs)
		stealth.ThinkTime(cfg.Humanization.ThinkTimeMs)
	}
	return nil
}

func extractName(page *rod.Page) string {
	if h1, err := page.Element("h1"); err == nil {
		txt, _ := h1.Text()
		return strings.TrimSpace(txt)
	}
	return ""
}

func extractCompany(page *rod.Page) string {
	if el, err := page.Element(`div.text-body-medium`); err == nil {
		txt, _ := el.Text()
		return strings.TrimSpace(txt)
	}
	return ""
}



