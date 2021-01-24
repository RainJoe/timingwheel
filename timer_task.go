package timingwheel

import (
	"time"
)

type TimerTask struct {
	expiration time.Time
	Do         func()
}
