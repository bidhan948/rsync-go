package progress

import (
	"fmt"
	"sync"
	"time"
)

type Tracker struct {
	mu        sync.Mutex
	total     int64
	current   int64
	lastPrint time.Time
}

func NewTracker(total int64) *Tracker {
	return &Tracker{
		total:     total,
		lastPrint: time.Now(),
	}
}

func (t *Tracker) Add(n int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current += n
	now := time.Now()
	if now.Sub(t.lastPrint) < time.Second {
		return
	}
	t.lastPrint = now
	var pct float64
	if t.total > 0 {
		pct = float64(t.current) * 100 / float64(t.total)
	}
	fmt.Printf("\r%.2f%% %d/%d bytes", pct, t.current, t.total)
}

func (t *Tracker) Done() {
	fmt.Println()
}
