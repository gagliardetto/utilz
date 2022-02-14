package utilz

import (
	"sync"
)

type ExitLight struct {
	mu        *sync.RWMutex
	isExiting bool
}

func NewExitLight() *ExitLight {
	return &ExitLight{
		mu: &sync.RWMutex{},
	}
}

func (light *ExitLight) IsExiting() bool {
	light.mu.RLock()
	defer light.mu.RUnlock()
	return light.isExiting
}

func (light *ExitLight) SetAsExiting() {
	light.mu.Lock()
	defer light.mu.Unlock()
	light.isExiting = true
}
