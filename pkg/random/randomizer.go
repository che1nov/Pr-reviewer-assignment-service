package random

import "math/rand"

type Adapter struct {
	rnd *rand.Rand
}

func New(rnd *rand.Rand) *Adapter {
	return &Adapter{rnd: rnd}
}

func (a *Adapter) Shuffle(n int, swap func(i, j int)) {
	if a == nil || a.rnd == nil || n < 2 {
		return
	}
	a.rnd.Shuffle(n, swap)
}
