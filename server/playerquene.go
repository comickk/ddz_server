package server

import (
	"sync"
)

//PlayerQueue
type PlayerQueue struct {
	m       map[string]*Player
	rwmutex sync.RWMutex
}

func (this *PlayerQueue) Get(key string) *Player {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()
	return this.m[key]
}

func (this *PlayerQueue) Put(key string, elem *Player) (*Player, bool) {
	this.rwmutex.Lock()
	defer this.rwmutex.Unlock()
	oldElem := this.m[key]
	this.m[key] = elem
	return oldElem, true
}

func (this *PlayerQueue) Remove(key string) *Player {
	this.rwmutex.Lock()
	defer this.rwmutex.Unlock()
	oldElem := this.m[key]
	delete(this.m, key)
	return oldElem
}

func (this *PlayerQueue) Clear() {
	this.rwmutex.Lock()
	defer this.rwmutex.Unlock()
	this.m = make(map[string]*Player)
}

func (this *PlayerQueue) Len() int {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()
	return len(this.m)
}

func (this *PlayerQueue) Contains(key string) bool {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()
	_, ok := this.m[key]
	return ok
}

func (this *PlayerQueue) Keys() []string {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()
	initialLen := len(this.m)
	keys := make([]string, initialLen)
	index := 0
	for k, _ := range this.m {
		keys[index] = k
		index++
	}
	return keys
}

func (this *PlayerQueue) Elems() []*Player {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()
	initialLen := len(this.m)
	elems := make([]*Player, initialLen)
	index := 0
	for _, v := range this.m {
		elems[index] = v
		index++
	}
	return elems
}

//取出三个玩家
func (this *PlayerQueue) Pop3() []*Player {
	this.rwmutex.RLock()
	defer this.rwmutex.RUnlock()

	initialLen := len(this.m)

	if initialLen > 2 {
		elems := make([]*Player, 3)
		index := 0
		for k, v := range this.m {
			elems[index] = v

			delete(this.m, k)

			index++
			if index == 3 {
				break
			}
		}
		return elems
	}

	return nil
}

func NewPlayerQueue() *PlayerQueue {
	return &PlayerQueue{m: make(map[string]*Player)}
}
