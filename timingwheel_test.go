package timingwheel

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func TestTimingWheel_DoAfter(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 60, time.Now(), workqueue.NewDelayingQueue())
	tw.DoAfter(func() {
		fmt.Println("Do task 2")
	}, 30*time.Millisecond)
	tw.DoAfter(func() {
		fmt.Println("Do task 1")
	}, 10*time.Millisecond)
	tw.DoAfter(func() {
		fmt.Println("Do task 3")
	}, 10*time.Second)
	go tw.Start()
	time.Sleep(10 * time.Second)
	tw.GracefulStop()
}
