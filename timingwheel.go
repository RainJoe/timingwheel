package timingwheel

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

type TimingWheel struct {
	buckets       []*TimerTaskList
	start         time.Time
	currentTime   time.Time
	tick          time.Duration
	size          int64
	interval      time.Duration
	delayQueue    workqueue.DelayingInterface
	overflowWheel *TimingWheel
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewTimingWheel(tick time.Duration, size int64, start time.Time, delayQueue workqueue.DelayingInterface) *TimingWheel {
	tw := &TimingWheel{
		buckets:     make([]*TimerTaskList, size),
		start:       start,
		currentTime: start.Truncate(tick),
		tick:        tick,
		size:        size,
		interval:    tick * time.Duration(size),
		delayQueue:  delayQueue,
		stopCh:      make(chan struct{}),
	}
	for i := range tw.buckets {
		t := &TimerTaskList{}
		t.Init()
		tw.buckets[i] = t
	}
	return tw
}

func (tw *TimingWheel) addOverflowWheel() {
	if tw.overflowWheel == nil {
		tw.overflowWheel = NewTimingWheel(tw.interval, tw.size, tw.currentTime, tw.delayQueue)
	}
}

func (tw *TimingWheel) add(task *TimerTask) bool {
	if task.expiration.Before(tw.currentTime.Add(tw.tick)) {
		return false
	}
	if task.expiration.Before(tw.currentTime.Add(tw.interval)) {
		virtualId := task.expiration.UnixNano() / tw.tick.Nanoseconds()
		bucket := tw.buckets[virtualId%tw.size]
		bucket.PushBack(task)
		if bucket.setExpiration(task.expiration.Truncate(tw.tick)) {
			delay := bucket.expiration.Sub(time.Now())
			tw.delayQueue.AddAfter(bucket, delay)
		}
	} else {
		if tw.overflowWheel == nil {
			tw.addOverflowWheel()
		}
		tw.overflowWheel.add(task)
	}
	return true
}

func (tw *TimingWheel) DoAfter(taskFunc func(), d time.Duration) {
	task := TimerTask{
		expiration: tw.currentTime.Add(d),
		Do:         taskFunc,
	}
	tw.add(&task)
}

func (tw *TimingWheel) advanceClock(t time.Time) {
	if t.After(tw.currentTime.Add(tw.tick)) || t.Equal(tw.currentTime.Add(tw.tick)) {
		tw.currentTime = t.Truncate(tw.tick)
	}
	if tw.overflowWheel != nil {
		tw.overflowWheel.advanceClock(tw.currentTime)
	}
}

func doTask(wg *sync.WaitGroup, task *TimerTask) {
	task.Do()
	wg.Done()
}

func (tw *TimingWheel) foreach(tasks *TimerTaskList) {
	tail := tasks.Back().Next()
	for i := tasks.Front(); i != tail; i = i.Next() {
		if i.Value != nil {
			task, _ := i.Value.(*TimerTask)
			tasks.Remove(i)
			if !tw.add(task) {
				tw.wg.Add(1)
				go doTask(&tw.wg, task)
			}
		}
	}
}

func (tw *TimingWheel) Start() {
	for {
		select {
		case <-tw.stopCh:
			return
		default:
			item, shutdown := tw.delayQueue.Get()
			if shutdown {
				break
			}
			tw.delayQueue.Done(item)
			tasks, _ := item.(*TimerTaskList)
			tw.advanceClock(tasks.expiration)
			tw.foreach(tasks)
		}
	}
}

func (tw *TimingWheel) Stop() {
	close(tw.stopCh)
}

func (tw *TimingWheel) GracefulStop() {
	tw.wg.Wait()
	close(tw.stopCh)
}
