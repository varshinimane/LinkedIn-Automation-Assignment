package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// TypeHumanWithTypos types text with occasional backspaces to mimic humans.
func TypeHumanWithTypos(el *rod.Element, text string) error {
	for _, ch := range text {
		if err := el.Input(string(ch)); err != nil {
			return err
		}
		time.Sleep(time.Duration(60+rand.Intn(120)) * time.Millisecond)
		// 10% chance of typo correction.
		if rand.Intn(10) == 0 {
			_ = el.Input("\b")
			time.Sleep(time.Duration(80+rand.Intn(120)) * time.Millisecond)
			_ = el.Input(string(ch))
		}
	}
	return nil
}

