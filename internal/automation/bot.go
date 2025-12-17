package automation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation/internal/behavior"
	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/config"
	"linkedin-automation/internal/state"
)

const (
	counterConnections = "connections_today"
	counterMessages    = "messages_today"
)

// Bot orchestrates browser actions and state.
type Bot struct {
	cfg       config.Config
	session   *browser.Session
	humanizer *behavior.Humanizer
	store     *state.Store
}

func NewBot(cfg config.Config, sess *browser.Session, store *state.Store) *Bot {
	return &Bot{
		cfg:       cfg,
		session:   sess,
		humanizer: behavior.NewHumanizer(cfg),
		store:     store,
	}
}

// Run executes a single automation cycle.
func (b *Bot) Run(ctx context.Context) error {
	if b.session == nil || b.session.Page == nil {
		return errors.New("session not initialized")
	}

	if err := b.ensureLoggedIn(ctx); err != nil {
		return err
	}

	// Rotate through search queries and attempt small batches.
	for _, query := range b.cfg.Behavior.TargetSearchQueries {
		if err := b.searchAndConnect(ctx, query); err != nil {
			log.Printf("warn: search %q: %v", query, err)
		}
		// Respect pacing and limits.
		if reachedLimit, _ := b.reachedDailyLimit(counterConnections, b.cfg.Behavior.DailyConnectionLimit); reachedLimit {
			log.Printf("connection quota reached for today")
			break
		}
	}
	return nil
}

func (b *Bot) ensureLoggedIn(ctx context.Context) error {
	page := b.session.Page

	if strings.Contains(page.MustInfo().URL, "feed") {
		return nil
	}

	if err := page.Navigate(b.cfg.BaseURL + "/login"); err != nil {
		return fmt.Errorf("navigate to login: %w", err)
	}
	b.humanizer.Pause()

	emailInput, err := page.Timeout(12 * time.Second).Element(`input#username`)
	if err != nil {
		return fmt.Errorf("email input: %w", err)
	}
	passInput, err := page.Timeout(12 * time.Second).Element(`input#password`)
	if err != nil {
		return fmt.Errorf("password input: %w", err)
	}

	if err := emailInput.SelectAllText(); err == nil {
		_ = emailInput.Input("")
	}
	if err := b.humanizer.TypeHumanized(emailInput, b.cfg.Credentials.Email); err != nil {
		return err
	}
	b.humanizer.Pause()
	if err := passInput.SelectAllText(); err == nil {
		_ = passInput.Input("")
	}
	if err := b.humanizer.TypeHumanized(passInput, b.cfg.Credentials.Password); err != nil {
		return err
	}

	btn, err := page.Element(`button[type="submit"]`)
	if err != nil {
		return fmt.Errorf("login button: %w", err)
	}
	if err := b.humanizer.ClickWithIntent(btn); err != nil {
		return fmt.Errorf("click login: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := page.Context(waitCtx).WaitNavigation(proto.PageLifecycleEventNameNetworkAlmostIdle); err != nil {
		return fmt.Errorf("post login wait: %w", err)
	}

	_ = b.store.SaveMetadata(state.Metadata{LastLoginAt: time.Now()})
	return nil
}

func (b *Bot) searchAndConnect(ctx context.Context, query string) error {
	page := b.session.Page
	if err := page.Navigate(b.cfg.BaseURL + "/search/results/people/?keywords=" + urlEncode(query)); err != nil {
		return err
	}
	b.humanizer.Scroll(page)

	cards, err := page.Elements(`button:has-text("Connect")`)
	if err != nil {
		return fmt.Errorf("find connect buttons: %w", err)
	}

	for _, c := range cards {
		if reached, _ := b.reachedDailyLimit(counterConnections, b.cfg.Behavior.DailyConnectionLimit); reached {
			return nil
		}

		if err := b.humanizer.ClickWithIntent(c); err != nil {
			continue
		}
		time.Sleep(500 * time.Millisecond)

		sendBtn, err := page.Element(`button:has-text("Send")`)
		if err == nil {
			_ = b.humanizer.ClickWithIntent(sendBtn)
		}

		if _, err := b.store.IncrementCounter(datedCounter(counterConnections)); err != nil {
			log.Printf("warn: increment connections: %v", err)
		}
		b.humanizer.Pause()
	}
	return nil
}

func (b *Bot) reachedDailyLimit(counter string, limit int) (bool, int) {
	key := datedCounter(counter)
	val, err := b.store.GetCounter(key)
	if err != nil {
		log.Printf("warn: counter increment %s: %v", counter, err)
		return false, 0
	}
	if val >= limit {
		return true, val
	}
	return false, val
}

func datedCounter(base string) string {
	return fmt.Sprintf("%s:%s", base, time.Now().Format("2006-01-02"))
}

func urlEncode(s string) string {
	replacer := strings.NewReplacer(" ", "%20")
	return replacer.Replace(s)
}

