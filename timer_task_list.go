package timingwheel

import (
	"container/list"
	"time"
)

type TimerTaskList struct {
	list.List
	expiration time.Time
}

func (tt *TimerTaskList) setExpiration(t time.Time) bool {
	if tt.expiration != t {
		tt.expiration = t
		return true
	}
	return false
}
