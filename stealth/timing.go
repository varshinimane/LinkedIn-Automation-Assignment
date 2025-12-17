package stealth

import (
	"math/rand"
	"time"
)

// RandomPause waits between min and max with jitter.
func RandomPause(minMs, maxMs, jitterMs int) {
	if maxMs <= minMs {
		maxMs = minMs + 50
	}
	delay := minMs + rand.Intn(maxMs-minMs)
	delay += rand.Intn(jitterMs + 1)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// ThinkTime simulates user contemplation pauses.
func ThinkTime(baseMs int) {
	extra := rand.Intn(baseMs/2 + 200)
	time.Sleep(time.Duration(baseMs+extra) * time.Millisecond)
}

// Cooldown enforces cooldown periods (e.g., after rate limit).
func Cooldown(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}



