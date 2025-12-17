package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// RandomScroll simulates human scrolling with pauses and reversals.
func RandomScroll(page *rod.Page) {
	count := 3 + rand.Intn(4)
	for i := 0; i < count; i++ {
		delta := float64(400 + rand.Intn(400))
		if i%2 == 0 && rand.Intn(2) == 0 {
			delta = -delta / 2
		}
		_ = page.Mouse.Scroll(0, delta, 10+rand.Intn(6))
		time.Sleep(time.Duration(600+rand.Intn(500)) * time.Millisecond)
	}
}



