package util

// Copyright (C) 2020 ConsenSys Software Inc

import (
	"time"
)

// Clock allows the time.Now to be mocked out for testing
type Clock interface {
	Now() time.Time
	//	After(d time.Duration) <-chan time.Time
}

var clock Clock

// GetTimeImpl returns the implementation of clock to use.
func GetTimeImpl() Clock {
	if clock == nil {
		SetRealClock()
	}
	return clock
}

// SetRealClock ensures the real clock is in use
func SetRealClock() {
	clock = newRealClock()
}

func newRealClock() (impl Clock) {
	r := realClock{}
	var _ Clock = &r // Enforce interface compliance
	return &r
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

//	func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

var mockedUnixTime int64

// SetMockedClock sets a fake timer
func SetMockedClock(fakeTime int64) {
	mockedUnixTime = fakeTime
	clock = newMockedClock()
}

func newMockedClock() (impl Clock) {
	r := mockedClock{}
	var _ Clock = &r // Enforce interface compliance
	return &r
}

type mockedClock struct{}

func (mockedClock) Now() time.Time {
	return time.Unix(mockedUnixTime, 0)
}
