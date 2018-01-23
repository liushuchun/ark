package sync

import (
	"sync"
)

type Semaphore struct {
	avail int
	mutex sync.Mutex
	cond  sync.Cond
}

func NewSemaphore(avail int) *Semaphore {

	p := &Semaphore{
		avail: avail,
	}
	p.cond.L = &p.mutex
	return p
}

func (p *Semaphore) Lock() {

	p.mutex.Lock()
	for p.avail == 0 {
		p.cond.Wait()
	}
	p.avail--
	p.mutex.Unlock()
}

func (p *Semaphore) Unlock() {

	p.mutex.Lock()
	p.avail++
	p.mutex.Unlock()

	p.cond.Signal()
}
