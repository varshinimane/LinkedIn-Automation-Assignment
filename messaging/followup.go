package messaging

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

// SendFollowups detects newly accepted connection requests, navigates to each profile,
// sends personalized follow-up messages using templates with variable substitution,
// and enforces daily message limits. Tracks sent messages to avoid duplicates.
// Returns an error if messaging fails critically.
func SendFollowups(ctx context.Context, page *rod.Page, cfg config.Config, store *storage.Store, log *logger.Logger) error {
	if err := detectAccepted(ctx, page, store, log); err != nil {
		return err
	}

	pending, err := store.PendingFollowups()
	if err != nil {
		return err
	}

	current, _ := store.GetDailyCounter("messages")
	if current >= cfg.Limits.DailyMessages {
		log.Warn("daily message limit reached", "count", current)
		return nil
	}

	for _, profileURL := range pending {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := page.Navigate(profileURL); err != nil {
			log.Warn("navigate profile failed", "url", profileURL, "err", err)
			continue
		}

		msgBtn, err := page.Timeout(8 * time.Second).ElementR("button", "Message")
		if err != nil {
			continue
		}

		box := msgBtn.MustShape().Box()
		_ = stealth.MoveHumanlike(page, box.X+box.Width/2, box.Y+box.Height/2)
		_ = msgBtn.Click(proto.InputMouseButtonLeft, 1)
		time.Sleep(600 * time.Millisecond)

		editor, err := page.Timeout(5 * time.Second).Element(`div[role="textbox"]`)
		if err != nil {
			log.Warn("message editor not found", "url", profileURL)
			continue
		}
		name := extractName(page)
		company := extractCompany(page)
		body := utils.Template(utils.RandomChoice(cfg.Messaging.Templates), map[string]string{
			"name":    name,
			"company": company,
		})
		if err := stealth.TypeHumanWithTypos(editor, body); err != nil {
			log.Warn("typing message failed", "url", profileURL, "err", err)
			continue
		}
		if send, err := page.ElementR("button", "Send"); err == nil {
			_ = send.Click(proto.InputMouseButtonLeft, 1)
		}
		_, _ = store.IncrementDailyCounter("messages")
		_ = store.SaveMessage(profileURL, body)
		log.Info("follow-up sent", "url", profileURL)

		if c, _ := store.GetDailyCounter("messages"); c >= cfg.Limits.DailyMessages {
			log.Warn("daily message limit reached", "count", c)
			return nil
		}

		stealth.RandomPause(cfg.Humanization.MinDelayMs, cfg.Humanization.MaxDelayMs, cfg.Humanization.JitterMs)
	}
	return nil
}

// detectAccepted marks requested connections that now show a Message button.
func detectAccepted(ctx context.Context, page *rod.Page, store *storage.Store, log *logger.Logger) error {
	requested, err := store.RequestedConnections()
	if err != nil {
		return err
	}
	for _, profileURL := range requested {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := page.Navigate(profileURL); err != nil {
			continue
		}
		if _, err := page.Timeout(5 * time.Second).ElementR("button", "Message"); err == nil {
			_ = store.MarkAccepted(profileURL)
			log.Info("accepted connection detected", "url", profileURL)
		}
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



