package enet

import (
	"container/heap"
)

type enet_timer_callback func()
type enet_timer_item struct {
	weight   int64
	callback enet_timer_callback
	index    int // heap index
}
type priority_queue []*enet_timer_item
type enet_timer_queue struct{ *priority_queue }

// sort interface

func (self priority_queue) Len() int           { return len(self) }
func (self priority_queue) Less(i, j int) bool { return self[i].weight < self[j].weight }
func (self priority_queue) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
	self[i].index = i
	self[j].index = j
}

// heap interface

func (self *priority_queue) Push(x interface{}) {
	v := x.(*enet_timer_item)
	v.index = len(*self)
	*self = append(*self, v)
}

func (self *priority_queue) Pop() interface{} {
	l := len(*self)
	v := (*self)[l-1]
	*self = (*self)[:l-1]
	v.index = -1
	return v
}

// timer queue interface
func new_enet_timer_queue() enet_timer_queue {
	timers := make(priority_queue, 0)
	heap.Init(&timers)
	return enet_timer_queue{&timers}
}
func (timers enet_timer_queue) push(deadline int64, cb enet_timer_callback) *enet_timer_item {
	v := &enet_timer_item{deadline, cb, -1}
	heap.Push(timers, v)
	return v
}

func (timers enet_timer_queue) pop(now int64) enet_timer_callback {
	if timers.Len() == 0 {
		return nil
	}
	if (*timers.priority_queue)[0].weight < now {
		top := heap.Pop(timers).(*enet_timer_item)
		return top.callback
	}
	return nil
}

func (timers enet_timer_queue) remove(idx int) {
	assert(idx < timers.Len())
	heap.Remove(timers, idx)
}
