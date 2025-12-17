package behavior

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation/internal/config"
)

// Humanizer applies jitter, scrolling, and minor actions to mimic users.
type Humanizer struct {
	cfg config.Config
	rng *rand.Rand
}

func NewHumanizer(cfg config.Config) *Humanizer {
	return &Humanizer{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Pause waits a random human-like delay.
func (h *Humanizer) Pause() {
	delay := h.cfg.Humanization.MinDelayMs + h.rng.Intn(max(1, h.cfg.Humanization.MaxDelayMs-h.cfg.Humanization.MinDelayMs))
	delay += h.rng.Intn(h.cfg.Humanization.JitterMs + 1)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// Scroll gently through the page.
func (h *Humanizer) Scroll(page *rod.Page) {
	for i := 0; i < 3+h.rng.Intn(3); i++ {
		_ = page.Mouse.Scroll(0, float64(300+h.rng.Intn(400)), 12+h.rng.Intn(4))
		time.Sleep(time.Duration(h.cfg.Humanization.ScrollPauseMs+h.rng.Intn(400)) * time.Millisecond)
	}
}

// Wiggle moves the mouse subtly to avoid static cursor.
func (h *Humanizer) Wiggle(page *rod.Page) {
	for i := 0; i < 4+h.rng.Intn(3); i++ {
		x := float64(100 + h.rng.Intn(600))
		y := float64(200 + h.rng.Intn(400))
		_ = proto.InputDispatchMouseEvent{
			Type: proto.InputDispatchMouseEventTypeMouseMoved,
			X:    x,
			Y:    y,
		}.Call(page)
		time.Sleep(time.Duration(80+h.rng.Intn(120)) * time.Millisecond)
	}
}

// ClickWithIntent adds a short hover before clicking.
func (h *Humanizer) ClickWithIntent(el *rod.Element) error {
	box := el.MustShape().Box()
	x := box.X + box.Width/2 + float64(h.rng.Intn(6)-3)
	y := box.Y + box.Height/2 + float64(h.rng.Intn(6)-3)
	h.Pause()
	page := el.Page()
	_ = proto.InputDispatchMouseEvent{
		Type: proto.InputDispatchMouseEventTypeMouseMoved,
		X:    x,
		Y:    y,
	}.Call(page)
	time.Sleep(time.Duration(150+h.rng.Intn(200)) * time.Millisecond)
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// TypeHumanized types with per-character variance.
func (h *Humanizer) TypeHumanized(el *rod.Element, text string) error {
	typingVariance := 80 // default variance in ms if not in config
	for _, ch := range text {
		if err := el.Input(string(ch)); err != nil {
			return err
		}
		time.Sleep(time.Duration(40+h.rng.Intn(typingVariance+1)) * time.Millisecond)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

