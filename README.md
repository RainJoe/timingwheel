# A simple timing wheel timer in go

## Introduction
Hierarchical timing wheel based on kafka timing wheel.

## Usage
```go
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
```
