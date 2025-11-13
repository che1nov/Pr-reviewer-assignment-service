package clock

import "time"

type Clock interface {
	Now() time.Time
}

type System struct{}

func NewSystem() *System {
	return &System{}
}

func (System) Now() time.Time {
	return time.Now().UTC()
}
