package health

import (
	"sync/atomic"
)

type Health struct {
	isReady atomic.Bool // global readiness flag, initial is false
}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) SetReady(ready bool) {
	h.isReady.Store(ready)
}
