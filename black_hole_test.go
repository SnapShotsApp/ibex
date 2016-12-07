package main

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func shouldNotTimeOutIn(actual interface{}, expected ...interface{}) string {
	fail := make(chan bool)
	out := make(chan bool)
	duration := expected[0].(time.Duration)
	timer := time.AfterFunc(duration, func() { fail <- true })
	go func() {
		actual.(func())()
		timer.Stop()
		out <- true
	}()

	select {
	case <-fail:
		return fmt.Sprintf("Expected function to return within %v", expected)
	case <-out:
		return ""
	}
}

func TestBlackHole(t *testing.T) {
	Convey("BlackHole chan accepts any number of things without blocking", t, func() {
		bh := NewBlackHole()
		f := func() {
			for i := 0; i <= 50; i++ {
				bh <- &stat{}
			}
		}

		So(f, shouldNotTimeOutIn, 1*time.Second)
	})
}
