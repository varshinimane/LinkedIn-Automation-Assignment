package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// MoveHumanlike moves the mouse along a curved path with overshoot and micro-corrections.
func MoveHumanlike(page *rod.Page, x, y float64) error {
	// Start from a slightly random origin near the top area to avoid relying on Position().
	startX := rand.Float64()*200 + 100
	startY := rand.Float64()*150 + 80

	// Overshoot slightly then correct.
	overshootX := x + rand.Float64()*12-6
	overshootY := y + rand.Float64()*12-6

	steps := 12 + rand.Intn(10)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		// quadratic Bezier for gentle curve
		ctrlX := (startX + overshootX) / 2
		ctrlY := startY + (overshootY-startY)*0.6
		curX := (1-t)*(1-t)*startX + 2*(1-t)*t*ctrlX + t*t*overshootX
		curY := (1-t)*(1-t)*startY + 2*(1-t)*t*ctrlY + t*t*overshootY
		if err := moveMouse(page, curX, curY); err != nil {
			return err
		}
		time.Sleep(variableDelay(12, 16))
	}

	// Micro-corrections around target.
	for i := 0; i < 3; i++ {
		mx := x + rand.Float64()*3-1.5
		my := y + rand.Float64()*3-1.5
		if err := moveMouse(page, mx, my); err != nil {
			return err
		}
		time.Sleep(variableDelay(20, 30))
	}

	// Final settle on target.
	return moveMouse(page, x, y)
}

// WanderCursor drifts cursor around the viewport to avoid stillness.
func WanderCursor(page *rod.Page, width, height float64) {
	for i := 0; i < 5+rand.Intn(4); i++ {
		x := rand.Float64()*width*0.8 + width*0.1
		y := rand.Float64()*height*0.8 + height*0.1
		_ = moveMouse(page, x, y)
		time.Sleep(variableDelay(80, 160))
	}
}

func variableDelay(minMs, maxMs int) time.Duration {
	ms := minMs + rand.Intn(int(math.Max(1, float64(maxMs-minMs))))
	return time.Duration(ms) * time.Millisecond
}

func moveMouse(page *rod.Page, x, y float64) error {
	return proto.InputDispatchMouseEvent{
		Type: proto.InputDispatchMouseEventTypeMouseMoved,
		X:    x,
		Y:    y,
	}.Call(page)
}

